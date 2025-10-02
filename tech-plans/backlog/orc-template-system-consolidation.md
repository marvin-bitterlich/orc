# ORC Template System Consolidation

**Status**: done

## Problem & Solution
**Current Issue:** Tech plan template was duplicated in multiple commands, and investigation CLAUDE.md template created dirty repository states while duplicating functionality already provided by /bootstrap and /janitor.
**Solution:** Centralized tech plan template in ~/src/orc/templates/ and eliminated investigation CLAUDE.md pattern entirely in favor of clean /bootstrap + /janitor workflow.

## Current System State
Tech plan template successfully extracted from embedded format in tech-plan command to centralized location. Investigation CLAUDE.md template pattern eliminated entirely - investigation Claude now uses /bootstrap + /janitor + .tech-plans/ for context. Repositories stay clean with no custom CLAUDE.md modifications.

## Implementation
### Approach
Extract tech plan template to centralized location and eliminate investigation CLAUDE.md template pattern entirely. Investigation Claude uses existing ORC commands (/bootstrap + /janitor) rather than custom context files.

### Simplified Investigation Workflow  
**Old Pattern**: Orchestrator creates custom CLAUDE.md in worktree → Investigation Claude reads custom context
**New Pattern**: Orchestrator creates .tech-plans/ → Investigation Claude runs /bootstrap + /janitor for context

### Template Loading System
Commands reference centralized template using `@/Users/looneym/src/orc/templates/tech-plan.md` notation for consistent structure across all tech plan creation.

## Testing Strategy
Validated template extraction and command updates work correctly. Template loads properly with @ notation.

## Implementation Plan
### Phase 1: Template Extraction ✅
- Created ~/src/orc/templates/ directory
- Extracted tech plan template to templates/tech-plan.md
- Updated tech-plan command to reference centralized template

### Phase 2: Architectural Simplification ✅  
- Eliminated investigation CLAUDE.md template pattern entirely
- Updated new-work command to use /bootstrap + /janitor workflow
- Removed orchestrator-workflow.md (replaced by new-work command)
- Repositories stay clean - no custom CLAUDE.md modifications

### Phase 3: Integration & Validation ✅
- Template loads correctly with @ notation
- Investigation handoff uses existing ORC commands
- Clean repository state maintained
- System ready for consistent tech plan creation with simplified investigation workflow