package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/core/cyclereceipt"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// CycleReceiptServiceImpl implements the CycleReceiptService interface.
type CycleReceiptServiceImpl struct {
	crecRepo secondary.CycleReceiptRepository
}

// NewCycleReceiptService creates a new CycleReceiptService with injected dependencies.
func NewCycleReceiptService(crecRepo secondary.CycleReceiptRepository) *CycleReceiptServiceImpl {
	return &CycleReceiptServiceImpl{
		crecRepo: crecRepo,
	}
}

// CreateCycleReceipt creates a new cycle receipt.
func (s *CycleReceiptServiceImpl) CreateCycleReceipt(ctx context.Context, req primary.CreateCycleReceiptRequest) (*primary.CreateCycleReceiptResponse, error) {
	// Validate CWO exists
	cwoExists, err := s.crecRepo.CWOExists(ctx, req.CWOID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate CWO: %w", err)
	}

	// Check if CWO already has a CREC (1:1 relationship)
	cwoHasCREC, err := s.crecRepo.CWOHasCREC(ctx, req.CWOID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing CREC: %w", err)
	}

	// Build guard context and evaluate
	guardCtx := cyclereceipt.CreateCRECContext{
		CWOID:            req.CWOID,
		CWOExists:        cwoExists,
		CWOHasCREC:       cwoHasCREC,
		DeliveredOutcome: req.DeliveredOutcome,
	}

	result := cyclereceipt.CanCreateCREC(guardCtx)
	if !result.Allowed {
		return nil, fmt.Errorf("%s", result.Reason)
	}

	// Get shipment ID from CWO
	shipmentID, err := s.crecRepo.GetCWOShipmentID(ctx, req.CWOID)
	if err != nil {
		return nil, fmt.Errorf("failed to get CWO shipment ID: %w", err)
	}

	// Get next ID
	nextID, err := s.crecRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CREC ID: %w", err)
	}

	// Create record
	record := &secondary.CycleReceiptRecord{
		ID:                nextID,
		CWOID:             req.CWOID,
		ShipmentID:        shipmentID,
		DeliveredOutcome:  req.DeliveredOutcome,
		Evidence:          req.Evidence,
		VerificationNotes: req.VerificationNotes,
		Status:            "draft",
	}

	if err := s.crecRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create CREC: %w", err)
	}

	// Fetch created CREC
	created, err := s.crecRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created CREC: %w", err)
	}

	return &primary.CreateCycleReceiptResponse{
		CycleReceiptID: created.ID,
		CycleReceipt:   s.recordToCREC(created),
	}, nil
}

// GetCycleReceipt retrieves a cycle receipt by ID.
func (s *CycleReceiptServiceImpl) GetCycleReceipt(ctx context.Context, crecID string) (*primary.CycleReceipt, error) {
	record, err := s.crecRepo.GetByID(ctx, crecID)
	if err != nil {
		return nil, err
	}
	return s.recordToCREC(record), nil
}

// GetCycleReceiptByCWO retrieves a cycle receipt by CWO ID.
func (s *CycleReceiptServiceImpl) GetCycleReceiptByCWO(ctx context.Context, cwoID string) (*primary.CycleReceipt, error) {
	record, err := s.crecRepo.GetByCWO(ctx, cwoID)
	if err != nil {
		return nil, err
	}
	return s.recordToCREC(record), nil
}

// ListCycleReceipts lists cycle receipts with optional filters.
func (s *CycleReceiptServiceImpl) ListCycleReceipts(ctx context.Context, filters primary.CycleReceiptFilters) ([]*primary.CycleReceipt, error) {
	records, err := s.crecRepo.List(ctx, secondary.CycleReceiptFilters{
		CWOID:      filters.CWOID,
		ShipmentID: filters.ShipmentID,
		Status:     filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list CRECs: %w", err)
	}

	crecs := make([]*primary.CycleReceipt, len(records))
	for i, r := range records {
		crecs[i] = s.recordToCREC(r)
	}
	return crecs, nil
}

// UpdateCycleReceipt updates a cycle receipt.
func (s *CycleReceiptServiceImpl) UpdateCycleReceipt(ctx context.Context, req primary.UpdateCycleReceiptRequest) error {
	// Verify CREC exists and is in draft status
	record, err := s.crecRepo.GetByID(ctx, req.CycleReceiptID)
	if err != nil {
		return err
	}

	if record.Status != "draft" {
		return fmt.Errorf("cannot update CREC %s: only draft CRECs can be updated (current status: %s)", req.CycleReceiptID, record.Status)
	}

	updateRecord := &secondary.CycleReceiptRecord{
		ID:                req.CycleReceiptID,
		DeliveredOutcome:  req.DeliveredOutcome,
		Evidence:          req.Evidence,
		VerificationNotes: req.VerificationNotes,
	}
	return s.crecRepo.Update(ctx, updateRecord)
}

// DeleteCycleReceipt deletes a cycle receipt.
func (s *CycleReceiptServiceImpl) DeleteCycleReceipt(ctx context.Context, crecID string) error {
	return s.crecRepo.Delete(ctx, crecID)
}

// SubmitCycleReceipt transitions a cycle receipt from draft to submitted.
func (s *CycleReceiptServiceImpl) SubmitCycleReceipt(ctx context.Context, crecID string) error {
	// Get current CREC
	record, err := s.crecRepo.GetByID(ctx, crecID)
	if err != nil {
		return err
	}

	// Check CWO exists and get its status
	cwoExists, err := s.crecRepo.CWOExists(ctx, record.CWOID)
	if err != nil {
		return fmt.Errorf("failed to validate CWO: %w", err)
	}

	var cwoStatus string
	if cwoExists {
		cwoStatus, err = s.crecRepo.GetCWOStatus(ctx, record.CWOID)
		if err != nil {
			return fmt.Errorf("failed to get CWO status: %w", err)
		}
	}

	// Build guard context and evaluate
	guardCtx := cyclereceipt.StatusTransitionContext{
		CRECID:        crecID,
		CurrentStatus: record.Status,
		CWOExists:     cwoExists,
		CWOStatus:     cwoStatus,
	}

	result := cyclereceipt.CanSubmit(guardCtx)
	if !result.Allowed {
		return fmt.Errorf("%s", result.Reason)
	}

	return s.crecRepo.UpdateStatus(ctx, crecID, "submitted")
}

// VerifyCycleReceipt transitions a cycle receipt from submitted to verified.
func (s *CycleReceiptServiceImpl) VerifyCycleReceipt(ctx context.Context, crecID string) error {
	// Get current CREC
	record, err := s.crecRepo.GetByID(ctx, crecID)
	if err != nil {
		return err
	}

	// Build guard context and evaluate
	guardCtx := cyclereceipt.StatusTransitionContext{
		CRECID:        crecID,
		CurrentStatus: record.Status,
	}

	result := cyclereceipt.CanVerify(guardCtx)
	if !result.Allowed {
		return fmt.Errorf("%s", result.Reason)
	}

	return s.crecRepo.UpdateStatus(ctx, crecID, "verified")
}

// Helper methods

func (s *CycleReceiptServiceImpl) recordToCREC(r *secondary.CycleReceiptRecord) *primary.CycleReceipt {
	return &primary.CycleReceipt{
		ID:                r.ID,
		CWOID:             r.CWOID,
		ShipmentID:        r.ShipmentID,
		DeliveredOutcome:  r.DeliveredOutcome,
		Evidence:          r.Evidence,
		VerificationNotes: r.VerificationNotes,
		Status:            r.Status,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

// Ensure CycleReceiptServiceImpl implements the interface
var _ primary.CycleReceiptService = (*CycleReceiptServiceImpl)(nil)
