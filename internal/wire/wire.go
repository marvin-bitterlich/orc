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
	"github.com/example/orc/internal/app"
	"github.com/example/orc/internal/db"
	"github.com/example/orc/internal/ports/primary"
)

var (
	missionService       primary.MissionService
	groveService         primary.GroveService
	shipmentService      primary.ShipmentService
	taskService          primary.TaskService
	noteService          primary.NoteService
	handoffService       primary.HandoffService
	tomeService          primary.TomeService
	conclaveService      primary.ConclaveService
	operationService     primary.OperationService
	investigationService primary.InvestigationService
	questionService      primary.QuestionService
	planService          primary.PlanService
	tagService           primary.TagService
	messageService       primary.MessageService
	once                 sync.Once
)

// MissionService returns the singleton MissionService instance.
func MissionService() primary.MissionService {
	once.Do(initServices)
	return missionService
}

// GroveService returns the singleton GroveService instance.
func GroveService() primary.GroveService {
	once.Do(initServices)
	return groveService
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

// initServices initializes all services and their dependencies.
// This is called once via sync.Once.
func initServices() {
	// Get database connection
	database, err := db.GetDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Create repository adapters (secondary ports) - sqlite adapters with injected DB
	missionRepo := sqlite.NewMissionRepository(database)
	groveRepo := sqlite.NewGroveRepository(database)
	agentProvider := persistence.NewAgentIdentityProvider()

	// Create effect executor with injected repositories
	executor := app.NewEffectExecutor(groveRepo, missionRepo)

	// Create services (primary ports implementation)
	missionService = app.NewMissionService(missionRepo, groveRepo, agentProvider, executor)
	groveService = app.NewGroveService(groveRepo, missionRepo, agentProvider, executor)

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
}

// MissionAdapter returns a new MissionAdapter writing to stdout.
// Each call creates a new adapter (adapters are stateless translators).
func MissionAdapter() *cliadapter.MissionAdapter {
	return MissionAdapterWithOutput(os.Stdout)
}

// MissionAdapterWithOutput returns a new MissionAdapter writing to the given output.
// This variant allows testing or alternate output destinations.
func MissionAdapterWithOutput(out io.Writer) *cliadapter.MissionAdapter {
	once.Do(initServices)
	return cliadapter.NewMissionAdapter(missionService, out)
}

// GroveAdapter returns a new GroveAdapter writing to stdout.
// Each call creates a new adapter (adapters are stateless translators).
func GroveAdapter() *cliadapter.GroveAdapter {
	return GroveAdapterWithOutput(os.Stdout)
}

// GroveAdapterWithOutput returns a new GroveAdapter writing to the given output.
// This variant allows testing or alternate output destinations.
func GroveAdapterWithOutput(out io.Writer) *cliadapter.GroveAdapter {
	once.Do(initServices)
	return cliadapter.NewGroveAdapter(groveService, out)
}
