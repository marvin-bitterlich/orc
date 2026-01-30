// Package wire provides dependency injection for the ORC application.
// It creates singleton services with lazy initialization.
package wire

import (
	"io"
	"log"
	"os"
	"sync"

	cliadapter "github.com/example/orc/internal/adapters/cli"
	"github.com/example/orc/internal/adapters/filesystem"
	"github.com/example/orc/internal/adapters/persistence"
	"github.com/example/orc/internal/adapters/sqlite"
	tmuxadapter "github.com/example/orc/internal/adapters/tmux"
	"github.com/example/orc/internal/app"
	"github.com/example/orc/internal/db"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

var (
	commissionService              primary.CommissionService
	shipmentService                primary.ShipmentService
	taskService                    primary.TaskService
	noteService                    primary.NoteService
	handoffService                 primary.HandoffService
	tomeService                    primary.TomeService
	conclaveService                primary.ConclaveService
	operationService               primary.OperationService
	planService                    primary.PlanService
	tagService                     primary.TagService
	messageService                 primary.MessageService
	repoService                    primary.RepoService
	prService                      primary.PRService
	factoryService                 primary.FactoryService
	workshopService                primary.WorkshopService
	workbenchService               primary.WorkbenchService
	receiptService                 primary.ReceiptService
	summaryService                 primary.SummaryService
	gatehouseService               primary.GatehouseService
	watchdogService                primary.WatchdogService
	approvalService                primary.ApprovalService
	escalationService              primary.EscalationService
	manifestService                primary.ManifestService
	commissionOrchestrationService *app.CommissionOrchestrationService
	tmuxService                    secondary.TMuxAdapter
	libraryRepo                    secondary.LibraryRepository
	shipyardRepo                   secondary.ShipyardRepository
	once                           sync.Once
)

// CommissionService returns the singleton CommissionService instance.
func CommissionService() primary.CommissionService {
	once.Do(initServices)
	return commissionService
}

// ShipmentService returns the singleton ShipmentService instance.
func ShipmentService() primary.ShipmentService {
	once.Do(initServices)
	return shipmentService
}

// TaskService returns the singleton TaskService instance.
func TaskService() primary.TaskService {
	once.Do(initServices)
	return taskService
}

// NoteService returns the singleton NoteService instance.
func NoteService() primary.NoteService {
	once.Do(initServices)
	return noteService
}

// HandoffService returns the singleton HandoffService instance.
func HandoffService() primary.HandoffService {
	once.Do(initServices)
	return handoffService
}

// TomeService returns the singleton TomeService instance.
func TomeService() primary.TomeService {
	once.Do(initServices)
	return tomeService
}

// ConclaveService returns the singleton ConclaveService instance.
func ConclaveService() primary.ConclaveService {
	once.Do(initServices)
	return conclaveService
}

// OperationService returns the singleton OperationService instance.
func OperationService() primary.OperationService {
	once.Do(initServices)
	return operationService
}

// PlanService returns the singleton PlanService instance.
func PlanService() primary.PlanService {
	once.Do(initServices)
	return planService
}

// TagService returns the singleton TagService instance.
func TagService() primary.TagService {
	once.Do(initServices)
	return tagService
}

// MessageService returns the singleton MessageService instance.
func MessageService() primary.MessageService {
	once.Do(initServices)
	return messageService
}

// RepoService returns the singleton RepoService instance.
func RepoService() primary.RepoService {
	once.Do(initServices)
	return repoService
}

// PRService returns the singleton PRService instance.
func PRService() primary.PRService {
	once.Do(initServices)
	return prService
}

// FactoryService returns the singleton FactoryService instance.
func FactoryService() primary.FactoryService {
	once.Do(initServices)
	return factoryService
}

// WorkshopService returns the singleton WorkshopService instance.
func WorkshopService() primary.WorkshopService {
	once.Do(initServices)
	return workshopService
}

// WorkbenchService returns the singleton WorkbenchService instance.
func WorkbenchService() primary.WorkbenchService {
	once.Do(initServices)
	return workbenchService
}

// ReceiptService returns the singleton ReceiptService instance.
func ReceiptService() primary.ReceiptService {
	once.Do(initServices)
	return receiptService
}

// SummaryService returns the singleton SummaryService instance.
func SummaryService() primary.SummaryService {
	once.Do(initServices)
	return summaryService
}

// GatehouseService returns the singleton GatehouseService instance.
func GatehouseService() primary.GatehouseService {
	once.Do(initServices)
	return gatehouseService
}

// WatchdogService returns the singleton WatchdogService instance.
func WatchdogService() primary.WatchdogService {
	once.Do(initServices)
	return watchdogService
}

// ApprovalService returns the singleton ApprovalService instance.
func ApprovalService() primary.ApprovalService {
	once.Do(initServices)
	return approvalService
}

// EscalationService returns the singleton EscalationService instance.
func EscalationService() primary.EscalationService {
	once.Do(initServices)
	return escalationService
}

// ManifestService returns the singleton ManifestService instance.
func ManifestService() primary.ManifestService {
	once.Do(initServices)
	return manifestService
}

// CommissionOrchestrationService returns the singleton CommissionOrchestrationService instance.
func CommissionOrchestrationService() *app.CommissionOrchestrationService {
	once.Do(initServices)
	return commissionOrchestrationService
}

// TMuxAdapter returns the singleton TMuxAdapter instance.
func TMuxAdapter() secondary.TMuxAdapter {
	once.Do(initServices)
	return tmuxService
}

// LibraryRepository returns the singleton LibraryRepository instance.
func LibraryRepository() secondary.LibraryRepository {
	once.Do(initServices)
	return libraryRepo
}

// ShipyardRepository returns the singleton ShipyardRepository instance.
func ShipyardRepository() secondary.ShipyardRepository {
	once.Do(initServices)
	return shipyardRepo
}

// initServices initializes all services and their dependencies.
// This is called once via sync.Once.
func initServices() {
	// Get database connection
	database, err := db.GetDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Create repository adapters (secondary ports) - sqlite adapters with injected DB
	commissionRepo := sqlite.NewCommissionRepository(database)
	agentProvider := persistence.NewAgentIdentityProvider()
	tmuxAdapter := tmuxadapter.NewAdapter()
	tmuxService = tmuxAdapter // Store for getter

	// Create workspace adapter (needed by effect executor and workshop service)
	home, _ := os.UserHomeDir()
	workspaceAdapter, err := filesystem.NewWorkspaceAdapter(home+"/wb", home+"/src") // ~/wb for worktrees, ~/src for repos
	if err != nil {
		log.Fatalf("failed to create workspace adapter: %v", err)
	}

	// Create effect executor with injected repositories and adapters
	executor := app.NewEffectExecutor(commissionRepo, tmuxAdapter, workspaceAdapter)

	// Create services (primary ports implementation)
	commissionService = app.NewCommissionService(commissionRepo, agentProvider, executor)

	// Create shipment and task services
	shipmentRepo := sqlite.NewShipmentRepository(database)
	taskRepo := sqlite.NewTaskRepository(database)
	tagRepo := sqlite.NewTagRepository(database)
	taskService = app.NewTaskService(taskRepo, tagRepo)

	// Create note, handoff, and tome services
	noteRepo := sqlite.NewNoteRepository(database)
	handoffRepo := sqlite.NewHandoffRepository(database)
	tomeRepo := sqlite.NewTomeRepository(database)
	noteService = app.NewNoteService(noteRepo)
	handoffService = app.NewHandoffService(handoffRepo)

	// Create library and shipyard repositories (used by services for park/unpark)
	libraryRepo = sqlite.NewLibraryRepository(database)
	shipyardRepo = sqlite.NewShipyardRepository(database)

	// Create tome and shipment services (need library/shipyard repos for park/unpark)
	tomeService = app.NewTomeService(tomeRepo, noteService, libraryRepo)
	shipmentService = app.NewShipmentService(shipmentRepo, taskRepo, shipyardRepo)

	// Create conclave and operation services
	conclaveRepo := sqlite.NewConclaveRepository(database)
	operationRepo := sqlite.NewOperationRepository(database)
	conclaveService = app.NewConclaveService(conclaveRepo)
	operationService = app.NewOperationService(operationRepo)

	// Create plan repository
	planRepo := sqlite.NewPlanRepository(database)

	// Create tag and message services
	tagService = app.NewTagService(tagRepo)
	messageRepo := sqlite.NewMessageRepository(database)
	messageService = app.NewMessageService(messageRepo)

	// Create repo and PR services
	repoRepo := sqlite.NewRepoRepository(database)
	prRepo := sqlite.NewPRRepository(database)
	repoService = app.NewRepoService(repoRepo)
	prService = app.NewPRService(prRepo, shipmentService)

	// Create factory, workshop, and workbench services
	factoryRepo := sqlite.NewFactoryRepository(database)
	workshopRepo := sqlite.NewWorkshopRepository(database)
	workbenchRepo := sqlite.NewWorkbenchRepository(database)
	gatehouseRepo := sqlite.NewGatehouseRepository(database) // needed by workshop service for auto-creation
	factoryService = app.NewFactoryService(factoryRepo)
	workshopService = app.NewWorkshopService(factoryRepo, workshopRepo, workbenchRepo, repoRepo, gatehouseRepo, tmuxService, workspaceAdapter, executor)
	workbenchService = app.NewWorkbenchService(workbenchRepo, workshopRepo, agentProvider, executor)

	// Create approval and escalation repositories early (needed by plan service)
	approvalRepo := sqlite.NewApprovalRepository(database)
	escalationRepo := sqlite.NewEscalationRepository(database)

	// Create plan service (needs multiple dependencies for escalation workflow)
	planService = app.NewPlanService(
		planRepo,
		approvalRepo,
		escalationRepo,
		workbenchRepo,
		gatehouseRepo,
		messageService,
		tmuxAdapter,
	)

	// Create receipt service
	receiptRepo := sqlite.NewReceiptRepository(database)
	receiptService = app.NewReceiptService(receiptRepo)

	// Create new entity services (gatehouses, watchdogs, approvals, escalations, manifests)
	// Pass workshop repo to gatehouse service for EnsureAllWorkshopsHaveGatehouses
	gatehouseService = app.NewGatehouseService(gatehouseRepo, workshopRepo)

	watchdogRepo := sqlite.NewWatchdogRepository(database)
	watchdogService = app.NewWatchdogService(watchdogRepo)

	approvalService = app.NewApprovalService(approvalRepo)

	escalationService = app.NewEscalationService(escalationRepo)

	manifestRepo := sqlite.NewManifestRepository(database)
	manifestService = app.NewManifestService(manifestRepo)

	// Create orchestration services
	commissionOrchestrationService = app.NewCommissionOrchestrationService(commissionService, agentProvider)

	// Create summary service (depends on most other services)
	summaryService = app.NewSummaryService(
		commissionService,
		conclaveService,
		tomeService,
		shipmentService,
		taskService,
		noteService,
		workbenchService,
		planService,
		approvalService,
		escalationService,
		receiptService,
	)
}

// ApplyGlobalTMuxBindings sets up ORC's global tmux key bindings.
// Safe to call repeatedly (idempotent). Silently ignores errors (tmux may not be running).
// This is called on every orc command invocation to ensure bindings are always current.
func ApplyGlobalTMuxBindings() {
	tmuxadapter.ApplyGlobalBindings()
}

// CommissionAdapter returns a new CommissionAdapter writing to stdout.
// Each call creates a new adapter (adapters are stateless translators).
func CommissionAdapter() *cliadapter.CommissionAdapter {
	return CommissionAdapterWithOutput(os.Stdout)
}

// CommissionAdapterWithOutput returns a new CommissionAdapter writing to the given output.
// This variant allows testing or alternate output destinations.
func CommissionAdapterWithOutput(out io.Writer) *cliadapter.CommissionAdapter {
	once.Do(initServices)
	return cliadapter.NewCommissionAdapter(commissionService, out)
}
