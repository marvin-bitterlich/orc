// Package wire provides dependency injection for the ORC application.
// It creates singleton services with lazy initialization.
package wire

import (
	"io"
	"log"
	"os"
	"sync"

	cliadapter "github.com/example/orc/internal/adapters/cli"
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
	investigationService           primary.InvestigationService
	questionService                primary.QuestionService
	planService                    primary.PlanService
	tagService                     primary.TagService
	messageService                 primary.MessageService
	repoService                    primary.RepoService
	prService                      primary.PRService
	factoryService                 primary.FactoryService
	workshopService                primary.WorkshopService
	workbenchService               primary.WorkbenchService
	commissionOrchestrationService *app.CommissionOrchestrationService
	tmuxService                    secondary.TMuxAdapter
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

// InvestigationService returns the singleton InvestigationService instance.
func InvestigationService() primary.InvestigationService {
	once.Do(initServices)
	return investigationService
}

// QuestionService returns the singleton QuestionService instance.
func QuestionService() primary.QuestionService {
	once.Do(initServices)
	return questionService
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

	// Create effect executor with injected repositories and adapters
	executor := app.NewEffectExecutor(commissionRepo, tmuxAdapter)

	// Create services (primary ports implementation)
	commissionService = app.NewCommissionService(commissionRepo, agentProvider, executor)

	// Create shipment and task services
	shipmentRepo := sqlite.NewShipmentRepository(database)
	taskRepo := sqlite.NewTaskRepository(database)
	tagRepo := sqlite.NewTagRepository(database)
	shipmentService = app.NewShipmentService(shipmentRepo, taskRepo)
	taskService = app.NewTaskService(taskRepo, tagRepo)

	// Create note, handoff, and tome services
	noteRepo := sqlite.NewNoteRepository(database)
	handoffRepo := sqlite.NewHandoffRepository(database)
	tomeRepo := sqlite.NewTomeRepository(database)
	noteService = app.NewNoteService(noteRepo)
	handoffService = app.NewHandoffService(handoffRepo)
	tomeService = app.NewTomeService(tomeRepo, noteService) // Tome needs NoteService for GetTomeNotes

	// Create conclave and operation services
	conclaveRepo := sqlite.NewConclaveRepository(database)
	operationRepo := sqlite.NewOperationRepository(database)
	conclaveService = app.NewConclaveService(conclaveRepo)
	operationService = app.NewOperationService(operationRepo)

	// Create investigation, question, and plan services
	investigationRepo := sqlite.NewInvestigationRepository(database)
	questionRepo := sqlite.NewQuestionRepository(database)
	planRepo := sqlite.NewPlanRepository(database)
	investigationService = app.NewInvestigationService(investigationRepo)
	questionService = app.NewQuestionService(questionRepo)
	planService = app.NewPlanService(planRepo)

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
	factoryService = app.NewFactoryService(factoryRepo)
	workshopService = app.NewWorkshopService(workshopRepo)
	workbenchService = app.NewWorkbenchService(workbenchRepo, workshopRepo, agentProvider, executor)

	// Create orchestration services
	commissionOrchestrationService = app.NewCommissionOrchestrationService(commissionService, agentProvider)
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
