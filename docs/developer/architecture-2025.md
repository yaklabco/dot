# Architecture Documentation (2025 Edition)

This document describes the technical architecture of dot, a type-safe symbolic link manager for configuration files.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Architectural Layers](#architectural-layers)
- [Design Principles](#design-principles)
- [Component Structure](#component-structure)
- [Data Flow](#data-flow)
- [Type System](#type-system)
- [Error Handling](#error-handling)
- [Concurrency Model](#concurrency-model)
- [Testing Strategy](#testing-strategy)
- [Dependency Rules](#dependency-rules)
- [Performance Characteristics](#performance-characteristics)
- [Security Considerations](#security-considerations)

## Architecture Overview

dot follows a layered architecture inspired by hexagonal architecture (ports and adapters) and functional programming principles. The system separates pure functional logic from side-effecting operations, enabling deterministic testing and safe execution.

### Core Architecture Pattern

The architecture implements the "Functional Core, Imperative Shell" pattern:

- **Functional Core**: Pure domain logic with no side effects (scanning, planning, resolution)
- **Imperative Shell**: Side-effecting operations isolated to executor layer (filesystem modifications)

This separation enables:
- Deterministic testing of core logic without filesystem access
- Safe rollback of failed operations
- Property-based testing of algebraic laws
- Parallelization of independent operations

## Architectural Layers

The system comprises six distinct layers, each with specific responsibilities and dependency constraints.

```mermaid
graph TB
    CLI[CLI Layer<br/>cmd/dot/]:::cliLayer
    API[API Layer<br/>pkg/dot/]:::apiLayer
    Pipeline[Pipeline Layer<br/>internal/pipeline/]:::pipelineLayer
    Executor[Executor Layer<br/>internal/executor/]:::executorLayer
    Core[Core Layer<br/>internal/scanner/<br/>internal/planner/<br/>internal/ignore/]:::coreLayer
    Domain[Domain Layer<br/>internal/domain/]:::domainLayer
    Adapters[Adapters<br/>internal/adapters/<br/>internal/manifest/<br/>internal/bootstrap/<br/>internal/config/]:::adaptersLayer
    
    CLI --> API
    API --> Pipeline
    API --> Executor
    API --> Adapters
    Pipeline --> Core
    Pipeline --> Adapters
    Executor --> Domain
    Core --> Domain
    Adapters --> Domain
    
    style CLI fill:#4A90E2,stroke:#2C5F8D,color:#fff
    style API fill:#50C878,stroke:#2D7A4A,color:#fff
    style Pipeline fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style Executor fill:#E67E22,stroke:#A84E0F,color:#fff
    style Core fill:#3498DB,stroke:#1F618D,color:#fff
    style Domain fill:#2ECC71,stroke:#1E8449,color:#fff
    style Adapters fill:#95A5A6,stroke:#5D6D7E,color:#fff
    
    classDef cliLayer stroke-width:3px
    classDef apiLayer stroke-width:3px
    classDef pipelineLayer stroke-width:3px
    classDef executorLayer stroke-width:3px
    classDef coreLayer stroke-width:3px
    classDef domainLayer stroke-width:3px
    classDef adaptersLayer stroke-width:2px,stroke-dasharray: 5 5
```

### 1. Domain Layer

**Location**: `internal/domain/`

**Purpose**: Pure domain model defining core types, operations, and port interfaces.

**Key Components**:
- Domain entities: `Package`, `Node`, `Plan`, `Operation`
- Phantom-typed paths: `PackagePath`, `TargetPath`, `FilePath`
- Port interfaces: `FS`, `Logger`, `Tracer`, `Metrics`
- Result types: `Result[T]` for monadic error handling
- Operation types (9 concrete implementations):
  - `LinkCreate`: Create symbolic link
  - `LinkDelete`: Remove symbolic link
  - `DirCreate`: Create directory
  - `DirDelete`: Remove empty directory
  - `DirRemoveAll`: Recursively remove directory
  - `FileMove`: Move file (with cross-device support)
  - `FileBackup`: Create backup copy
  - `FileDelete`: Delete file
  - `DirCopy`: Recursive directory copy
- Conflict representations
- Error types

**Characteristics**:
- No external dependencies except standard library
- All operations are data structures with execute/rollback methods
- Phantom types provide compile-time path safety
- Defines contracts (interfaces) for infrastructure

**Dependencies**: None (depends only on Go standard library)

### 2. Core Layer

**Location**: `internal/scanner/`, `internal/planner/`, `internal/ignore/`

**Purpose**: Pure functional logic for scanning packages, computing desired state, and planning operations.

**Key Components**:

**Scanner** (`internal/scanner/`):
- Package scanning with per-package ignore patterns
- Filesystem tree construction
- Dotfile name translation (e.g., `dot-bashrc` to `.bashrc`)
- Reserved package name detection
- Large file handling with interactive prompts
- Configuration: `ScanConfig` with `PerPackageIgnore`, `MaxFileSize`, `Interactive`

**Planner** (`internal/planner/`):
- Desired state computation
- Conflict detection and resolution
- Dependency graph construction
- Topological sorting for operation ordering
- Parallel execution batch computation
- Resolution policies: Fail, Backup, Overwrite
- Conflict resolution with backup directory support

**Ignore** (`internal/ignore/`):
- Pattern matching for file exclusion (glob-based)
- Default ignore patterns
- Custom pattern support
- IgnoreSet for efficient matching

**Characteristics**:
- Pure functions with no side effects
- Deterministic outputs for given inputs
- Testable without filesystem access
- Uses Result types for error handling

**Dependencies**: Domain layer only

### 3. Pipeline Layer

**Location**: `internal/pipeline/`

**Purpose**: Composable pipeline stages with generic type parameters for operation orchestration.

**Key Components**:
- `Pipeline[TIn, TOut]`: Generic pipeline type
- `ScanStage()`: Package scanning stage
- `PlanStage()`: Desired state computation stage
- `ResolveStage()`: Conflict resolution stage with targeted scanning
- `SortStage()`: Topological sorting stage
- `ManagePipeline`: Composition of stages for manage operations

**Characteristics**:
- Generic type parameters for type safety
- Composable stages using function composition
- Context-aware for cancellation support
- Monadic error propagation through stages
- Optimized scanning: only checks paths relevant to desired state (not full directory traversal)

**Pipeline Composition Example**:
```
ScanInput -> ScanStage -> []Package -> PlanStage -> DesiredState -> ResolveStage -> ResolveResult -> SortStage -> Plan
```

**Dependencies**: Domain and Core layers, Adapters

### 4. Executor Layer

**Location**: `internal/executor/`

**Purpose**: Transactional execution of plans with two-phase commit and automatic rollback.

**Key Components**:
- `Executor`: Main execution engine
- `CheckpointStore`: State checkpoint for rollback (memory-based default)
- Precondition validation with pending operation tracking
- Operation execution (sequential or parallel)
- Automatic rollback on failure
- Parallel execution support via batches

**Execution Phases**:

1. **Prepare Phase**: Validate all operations before execution, track pending creations
2. **Checkpoint Creation**: Save state for potential rollback
3. **Commit Phase**: Execute operations (sequential or parallel batches)
4. **Rollback Phase**: Undo operations if failures occur (reverse order)
5. **Checkpoint Cleanup**: Remove checkpoint on success

```mermaid
stateDiagram-v2
    [*] --> Prepare: Receive Plan
    
    Prepare --> ValidateOps: Validate Operations
    ValidateOps --> CheckPreconditions: Check Preconditions with Pending
    CheckPreconditions --> CreateCheckpoint: All Valid
    CheckPreconditions --> Failed: Validation Failed
    
    CreateCheckpoint --> CommitPhase: Checkpoint Saved
    CreateCheckpoint --> Failed: Checkpoint Failed
    
    CommitPhase --> ChooseMode: Choose Execution Mode
    ChooseMode --> ExecuteSequential: No Batches
    ChooseMode --> ExecuteParallel: Has Batches
    
    ExecuteSequential --> ExecuteOp1: Operation 1
    ExecuteOp1 --> ExecuteOp2: Success
    ExecuteOp1 --> Rollback: Operation Failed
    ExecuteOp2 --> UpdateComplete: All Ops Complete
    ExecuteOp2 --> Rollback: Operation Failed
    
    ExecuteParallel --> ExecuteBatch1: Batch 1 (Concurrent)
    ExecuteBatch1 --> ExecuteBatch2: Success
    ExecuteBatch1 --> Rollback: Operation Failed
    ExecuteBatch2 --> ExecuteBatch3: Success
    ExecuteBatch2 --> Rollback: Operation Failed
    ExecuteBatch3 --> UpdateComplete: All Batches Complete
    ExecuteBatch3 --> Rollback: Operation Failed
    
    UpdateComplete --> CleanupCheckpoint: Operations Done
    CleanupCheckpoint --> Success: Checkpoint Removed
    
    Rollback --> RestoreState: Undo Operations (Reverse Order)
    RestoreState --> RemoveCheckpoint: State Restored
    RemoveCheckpoint --> Failed: Rollback Complete
    
    Success --> [*]
    Failed --> [*]
    
    note right of Prepare
        Validate all operations
        Track pending directory/file creations
    end note
    
    note right of CommitPhase
        Execute in topologically
        sorted batches if available
    end note
    
    note right of Rollback
        Automatic rollback ensures
        no partial state
    end note
```

**Precondition Validation**: Checks source exists, parent directories exist, permissions, but also accounts for pending directory/file creations from earlier operations in the plan.

**Characteristics**:
- All-or-nothing transaction semantics
- Automatic rollback on any failure
- Support for parallel execution of independent operations
- Comprehensive error tracking
- Operation-level execute and rollback methods

**Dependencies**: Domain layer for types and ports

### 5. API Layer

**Location**: `pkg/dot/`

**Purpose**: Clean public Go library interface for embedding dot in other applications.

**Key Components**:
- `Client`: Facade delegating to specialized services
- `Config`: Configuration structure with validation
- **Service implementations**:
  - `ManageService`: Package installation with incremental hash-based remanage
  - `UnmanageService`: Package removal with restore support
  - `StatusService`: Status queries
  - `DoctorService`: Health checks with parallel scanning, triage, and pattern categorization
  - `AdoptService`: File adoption with directory flattening
  - `CloneService`: Git cloning with bootstrap and authentication
  - `BootstrapService`: Bootstrap configuration generation
  - `ManifestService`: State persistence and backup management

**Service Pattern**:

The Client uses a service-based architecture where each major operation is implemented by a dedicated service. This provides:
- Single Responsibility Principle adherence
- Independent testing of each service
- Clear boundaries between concerns
- Maintainable codebase

```mermaid
graph LR
    Client[Client<br/>Facade]:::clientNode
    
    ManageService[ManageService<br/>Package Installation]:::serviceNode
    UnmanageService[UnmanageService<br/>Package Removal]:::serviceNode
    StatusService[StatusService<br/>Status Queries]:::serviceNode
    DoctorService[DoctorService<br/>Health Checks]:::serviceNode
    AdoptService[AdoptService<br/>File Adoption]:::serviceNode
    CloneService[CloneService<br/>Repository Cloning]:::serviceNode
    BootstrapService[BootstrapService<br/>Config Generation]:::serviceNode
    ManifestService[ManifestService<br/>State Persistence]:::serviceNode
    
    Pipeline[Pipeline Layer]:::layerNode
    Executor[Executor Layer]:::layerNode
    Manifest[Manifest Store]:::layerNode
    
    Client --> ManageService
    Client --> UnmanageService
    Client --> StatusService
    Client --> DoctorService
    Client --> AdoptService
    Client --> CloneService
    Client --> BootstrapService
    
    ManageService --> Pipeline
    ManageService --> Executor
    ManageService --> ManifestService
    ManageService --> UnmanageService
    
    UnmanageService --> Executor
    UnmanageService --> ManifestService
    
    StatusService --> ManifestService
    DoctorService --> ManifestService
    DoctorService --> AdoptService
    
    AdoptService --> Executor
    AdoptService --> ManifestService
    
    CloneService --> ManageService
    CloneService --> ManifestService
    
    ManifestService --> Manifest
    
    style Client fill:#4A90E2,stroke:#2C5F8D,color:#fff,stroke-width:4px
    style ManageService fill:#50C878,stroke:#2D7A4A,color:#fff
    style UnmanageService fill:#E67E22,stroke:#A84E0F,color:#fff
    style StatusService fill:#3498DB,stroke:#1F618D,color:#fff
    style DoctorService fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style AdoptService fill:#1ABC9C,stroke:#148F77,color:#fff
    style CloneService fill:#F39C12,stroke:#B97A0F,color:#fff
    style BootstrapService fill:#E74C3C,stroke:#C0392B,color:#fff
    style ManifestService fill:#95A5A6,stroke:#7F8C8D,color:#fff
    style Pipeline fill:#34495E,stroke:#1C2833,color:#fff
    style Executor fill:#34495E,stroke:#1C2833,color:#fff
    style Manifest fill:#34495E,stroke:#1C2833,color:#fff
    
    classDef clientNode stroke-width:4px
    classDef serviceNode stroke-width:2px
    classDef layerNode stroke-width:2px,stroke-dasharray: 5 5
```

**Characteristics**:
- Stable public API
- Thread-safe operations
- Service-based delegation
- Comprehensive validation
- Each service is independently testable

**Dependencies**: All internal layers

### 6. CLI Layer

**Location**: `cmd/dot/`

**Purpose**: Cobra-based command-line interface providing user interaction.

**Key Components**:
- Command definitions (manage, unmanage, status, doctor, list, adopt, clone, remanage, upgrade)
- Flag parsing and validation
- Configuration loading from files and environment
- Output formatting (table, JSON, YAML)
- Progress indicators
- Error rendering with suggestions
- Golden file testing for output validation

**Characteristics**:
- Cobra command structure
- Viper configuration management
- Multiple output formats
- Rich error messages with context
- Interactive prompts for triage and adoption

**Dependencies**: API layer only (does not import internal packages directly)

## Design Principles

### Functional Core, Imperative Shell

Pure functional logic (scanning, planning, resolution) is separated from side-effecting operations (filesystem modifications). This enables:

- Deterministic testing without filesystem access
- Property-based testing of algebraic laws
- Safe parallelization
- Reliable rollback mechanisms

### Type Safety

Phantom types encode path semantics at compile time:

```go
type PackagePath struct { path string }
type TargetPath struct { path string }
type FilePath struct { path string }
```

This prevents path-related bugs:
- Cannot pass target path where package path expected
- Cannot mix relative and absolute paths incorrectly
- Compile-time validation of path operations

### Explicit Error Handling

The system uses `Result[T]` types for monadic error handling:

```go
type Result[T any] struct {
    value T
    err   error
    isOk  bool
}
```

This provides:
- No silent failures
- Explicit error propagation
- Composable error handling (Map, FlatMap, Collect)
- Type-safe success values

### Transactional Operations

All operations use two-phase commit:

1. **Validate**: Check preconditions (including pending operations)
2. **Execute**: Apply changes
3. **Rollback**: Undo on failure

This ensures:
- Atomic operation sets
- Automatic cleanup on failure
- No partial state on errors
- Safe concurrent execution

### Dependency Inversion

Infrastructure dependencies are abstracted through port interfaces:

```go
type FS interface {
    Stat(ctx context.Context, path string) (FileInfo, error)
    ReadDir(ctx context.Context, path string) ([]DirEntry, error)
    Symlink(ctx context.Context, oldname, newname string) error
    // ... other operations
}
```

This enables:
- Testing with memory-based implementations
- Platform-specific adapters
- Mock implementations for testing
- Isolation of domain logic from infrastructure

## Component Structure

### Adapter Pattern

The system uses adapters to implement port interfaces:

**Filesystem Adapters** (`internal/adapters/`):
- `OSFilesystem`: Production filesystem using `os` package (with memfs wrapper for go-billy compatibility)
- `MemFilesystem`: In-memory filesystem for testing
- `NoopFilesystem`: No-op implementation for dry-run mode

**Logging Adapters** (`internal/adapters/`):
- `SlogLogger`: Production logger using `log/slog`
- `NoopLogger`: Silent logger for testing

**Git Adapters** (`internal/adapters/`):
- `GoGitCloner`: Git cloning using go-git library
- Authentication: `NoAuth`, `TokenAuth`, `SSHAuth`

This pattern provides:
- Swappable implementations
- Testability without real filesystem
- Dry-run mode support
- Platform-specific optimizations

### Manifest Persistence

**Location**: `internal/manifest/`

**Purpose**: State persistence for tracking installed packages.

**Components**:
- `Manifest`: Package installation record
- `ManifestStore`: Interface for persistence
- `FSManifestStore`: File-based implementation
- `ContentHasher`: SHA256 hash computation for change detection

**Manifest Structure**:
```go
type Manifest struct {
    Version    string
    UpdatedAt  time.Time
    Packages   map[string]PackageInfo
    Hashes     map[string]string
    Repository *RepositoryInfo
    Doctor     *DoctorState
}

type PackageInfo struct {
    Name        string
    InstalledAt time.Time
    LinkCount   int
    Links       []string
    Backups     map[string]string // target -> backup path
    Source      PackageSource     // "managed" or "adopted"
    TargetDir   string
    PackageDir  string
}

type RepositoryInfo struct {
    URL       string
    Branch    string
    ClonedAt  time.Time
    CommitSHA string
}

type DoctorState struct {
    IgnoredLinks    map[string]IgnoredLink
    IgnoredPatterns []string
}
```

**Persistence Location**: `<TargetDir>/.dot-manifest.json`

**Purpose**:
- Track installed packages with source type (managed vs adopted)
- Enable incremental updates via content hashing
- Support status queries without filesystem scanning
- Facilitate safe uninstall operations
- Track repository information for cloned dotfiles
- Store doctor triage decisions (ignored links and patterns)
- Track backup locations for restoration

### Configuration System

**Location**: `internal/config/`

**Purpose**: Configuration loading, validation, and marshaling.

**Features**:
- Multiple format support (YAML, JSON, TOML)
- Precedence handling (flags > environment > files > defaults)
- XDG Base Directory Specification compliance
- Schema validation
- Default value application
- Writer with comment support

**Configuration Sources** (in precedence order):
1. Command-line flags
2. Environment variables (`DOT_*` prefix)
3. Project-local config (`./.dotrc`)
4. User config (`~/.config/dot/config.yaml`)
5. System config (`/etc/dot/config.yaml`)
6. Default values

### Bootstrap System

**Location**: `internal/bootstrap/`

**Purpose**: Repository setup configuration for automated package installation.

**Components**:
- `Config`: Bootstrap configuration schema
- `Loader`: Configuration file loading
- `Generator`: Bootstrap config generation from existing installation
- `PackageSpec`: Package definition with platform filtering
- `Profile`: Named package sets
- `Defaults`: Default conflict resolution policies

**Bootstrap Configuration**:
```yaml
version: "1.0"
packages:
  - name: vim
    required: true
    platform: [linux, darwin]
  - name: tmux
    platform: [linux]
profiles:
  minimal:
    description: Minimal installation
    packages: [vim]
  full:
    description: Full installation
    packages: [vim, tmux]
defaults:
  on_conflict: backup
  profile: minimal
```

**Features**:
- Platform-specific package filtering
- Profile-based installation
- Per-package conflict policies
- Required package enforcement
- Validation with error reporting

## Data Flow

### Manage Operation Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Layer
    participant API as API Client
    participant ManageSvc as ManageService
    participant Pipeline as Pipeline
    participant Scanner as Scanner
    participant Planner as Planner
    participant Executor as Executor
    participant FS as Filesystem
    participant ManifestSvc as ManifestService
    
    User->>CLI: dot manage vim tmux
    CLI->>CLI: Parse flags & load config
    CLI->>API: Client.Manage(ctx, packages)
    API->>ManageSvc: Manage(ctx, packages)
    
    ManageSvc->>ManageSvc: Validate package names (filter reserved)
    ManageSvc->>ManageSvc: PlanManage(ctx, packages)
    
    rect rgb(40, 70, 100)
        note right of ManageSvc: Pipeline Execution
        ManageSvc->>Pipeline: Execute(ScanInput)
        Pipeline->>Scanner: ScanPackage with config
        Scanner->>FS: Read package directories
        FS-->>Scanner: File tree
        Scanner-->>Pipeline: []Package
        
        Pipeline->>Planner: ComputeDesiredState
        Planner-->>Pipeline: DesiredState
        
        Pipeline->>Pipeline: scanCurrentState (targeted)
        Pipeline->>FS: Stat specific paths only
        FS-->>Pipeline: CurrentState
        
        Pipeline->>Planner: Resolve conflicts
        Planner-->>Pipeline: ResolveResult
        
        Pipeline->>Planner: TopologicalSort
        Planner-->>Pipeline: Sorted operations
        Pipeline-->>ManageSvc: Plan
    end
    
    ManageSvc->>ManageSvc: Check for conflicts
    
    rect rgb(80, 100, 50)
        note right of ManageSvc: Execution
        ManageSvc->>Executor: Execute(plan)
        Executor->>Executor: Prepare & validate
        Executor->>Executor: Create checkpoint
        
        alt Parallel execution
            loop For each batch
                Executor->>FS: Execute operations (concurrent)
            end
        else Sequential execution
            loop For each operation
                Executor->>FS: Execute operation
            end
        end
        
        alt Success
            Executor-->>ManageSvc: ExecutionResult
        else Failure
            Executor->>FS: Rollback changes
            Executor-->>ManageSvc: Error with rollback info
        end
    end
    
    ManageSvc->>ManifestSvc: Update(packages, plan)
    ManifestSvc->>ManifestSvc: Compute content hashes
    ManifestSvc->>FS: Write .dot-manifest.json
    ManageSvc-->>API: Success
    API-->>CLI: nil error
    CLI->>User: Display success message
```

### Remanage Operation Flow (Incremental)

```mermaid
sequenceDiagram
    participant User
    participant ManageSvc as ManageService
    participant ManifestSvc as ManifestService
    participant Hasher as ContentHasher
    participant FS as Filesystem
    participant Executor as Executor
    
    User->>ManageSvc: Remanage(vim)
    ManageSvc->>ManifestSvc: Load manifest
    ManifestSvc-->>ManageSvc: Manifest
    
    loop For each package
        ManageSvc->>ManageSvc: Check if in manifest
        
        alt Package not in manifest
            ManageSvc->>ManageSvc: PlanManage (full install)
        else Package in manifest
            ManageSvc->>Hasher: HashPackage(packagePath)
            Hasher->>FS: Read package files
            FS-->>Hasher: File contents
            Hasher-->>ManageSvc: Current hash
            
            ManageSvc->>ManageSvc: Compare with stored hash
            
            alt Hashes match
                ManageSvc->>ManageSvc: Verify links exist
                FS-->>ManageSvc: Link status
                
                alt All links exist
                    ManageSvc->>ManageSvc: No operations needed
                else Links missing
                    ManageSvc->>ManageSvc: PlanFullRemanage
                end
            else Hashes differ
                ManageSvc->>ManageSvc: PlanFullRemanage
                
                alt Adopted package
                    ManageSvc->>ManageSvc: Recreate single symlink
                else Managed package
                    ManageSvc->>ManageSvc: Unmanage + Manage
                    Note right of ManageSvc: Remove old links in memory<br/>before scanning to avoid skips
                end
            end
        end
    end
    
    ManageSvc->>Executor: Execute(plan)
    Executor-->>ManageSvc: Result
    ManageSvc->>ManifestSvc: UpdateWithSource
    ManifestSvc->>ManifestSvc: Preserve source type
```

### Doctor Health Check Flow with Triage

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Layer
    participant DoctorSvc as DoctorService
    participant ManifestSvc as ManifestService
    participant FS as Filesystem
    participant HealthChecker as HealthChecker
    participant Workers as Worker Pool
    participant Categorizer as Pattern Categorizer
    
    User->>CLI: dot doctor --triage
    CLI->>DoctorSvc: DoctorWithScan(ctx, scanCfg)
    
    DoctorSvc->>ManifestSvc: Load manifest
    ManifestSvc-->>DoctorSvc: Manifest
    
    rect rgb(50, 80, 120)
        note right of DoctorSvc: Check Managed Packages
        loop For each package
            loop For each link
                DoctorSvc->>HealthChecker: CheckLink
                HealthChecker->>FS: Lstat link
                HealthChecker->>FS: ReadLink
                HealthChecker->>FS: Stat target
                HealthChecker-->>DoctorSvc: HealthCheckResult
            end
        end
    end
    
    rect rgb(70, 90, 70)
        note right of DoctorSvc: Orphan Scan (Parallel)
        DoctorSvc->>DoctorSvc: Determine scan directories
        DoctorSvc->>DoctorSvc: Build ignore set from manifest
        DoctorSvc->>Workers: Start worker pool (NumCPU)
        
        par Parallel directory scanning
            Workers->>FS: Scan directory 1
            Workers->>FS: Scan directory 2
            Workers->>FS: Scan directory N
        end
        
        Workers-->>DoctorSvc: Aggregate results
        
        note right of DoctorSvc: Respects MaxIssues budget<br/>Early cancellation on limit
    end
    
    DoctorSvc-->>CLI: DiagnosticReport
    
    alt Triage mode
        CLI->>DoctorSvc: Triage(scanCfg, opts)
        
        rect rgb(100, 70, 80)
            note right of DoctorSvc: Categorization
            loop For each orphaned link
                DoctorSvc->>FS: ReadLink to get target
                DoctorSvc->>Categorizer: CategorizeSymlink
                Categorizer-->>DoctorSvc: Category (high/medium/low confidence)
            end
        end
        
        rect rgb(90, 70, 100)
            note right of DoctorSvc: Interactive Triage
            DoctorSvc->>User: Display groups by category
            User->>DoctorSvc: Choose processing mode (category/linear/auto)
            
            alt Category mode
                loop For each category
                    DoctorSvc->>User: Display category info
                    User->>DoctorSvc: Action (ignore/review/skip)
                    
                    alt Ignore
                        DoctorSvc->>ManifestSvc: AddIgnoredPattern
                    else Review
                        loop For each link
                            User->>DoctorSvc: Action per link
                        end
                    end
                end
            else Auto mode
                DoctorSvc->>ManifestSvc: Auto-ignore high confidence
            end
            
            DoctorSvc->>ManifestSvc: Save manifest
        end
    end
    
    CLI->>User: Display report/triage results
```

### Clone Operation Flow

```mermaid
sequenceDiagram
    participant User
    participant CloneSvc as CloneService
    participant Auth as AuthResolver
    participant GitCloner as GoGitCloner
    participant FS as Filesystem
    participant BootstrapLoader as Bootstrap Loader
    participant Selector as Package Selector
    participant ManageSvc as ManageService
    participant ManifestSvc as ManifestService
    
    User->>CloneSvc: Clone(repoURL, opts)
    
    CloneSvc->>FS: Validate packageDir is empty
    alt Not empty and not force
        FS-->>CloneSvc: Error
    end
    
    CloneSvc->>Auth: ResolveAuth(repoURL)
    Auth->>Auth: Check for SSH_AUTH_SOCK
    Auth->>Auth: Check for GITHUB_TOKEN
    Auth-->>CloneSvc: AuthMethod (SSH/Token/None)
    
    CloneSvc->>GitCloner: Clone(repoURL, packageDir, auth)
    GitCloner->>GitCloner: Shallow clone (depth=1)
    GitCloner-->>CloneSvc: Success
    
    CloneSvc->>BootstrapLoader: Load .dotbootstrap.yaml
    
    alt Bootstrap exists
        BootstrapLoader-->>CloneSvc: Config
        CloneSvc->>CloneSvc: FilterPackagesByPlatform
        CloneSvc->>CloneSvc: Filter reserved packages
        
        alt Profile specified
            CloneSvc->>CloneSvc: SelectPackagesFromProfile
        else Interactive or terminal interactive
            CloneSvc->>Selector: Select(packages)
            Selector->>User: Interactive selection UI
            User-->>Selector: Selected packages
            Selector-->>CloneSvc: Selected packages
        else Default profile exists
            CloneSvc->>CloneSvc: Use default profile
        else Non-interactive
            CloneSvc->>CloneSvc: Install all packages
        end
    else No bootstrap
        CloneSvc->>FS: Discover packages
        FS-->>CloneSvc: Package list
        
        alt Interactive
            CloneSvc->>Selector: Select(packages)
            Selector-->>CloneSvc: Selected packages
        else Non-interactive
            CloneSvc->>CloneSvc: Install all
        end
    end
    
    CloneSvc->>ManageSvc: Manage(packages)
    ManageSvc-->>CloneSvc: Success
    
    CloneSvc->>CloneSvc: Read .git/HEAD for branch/SHA
    CloneSvc->>ManifestSvc: UpdateRepository(repoInfo)
    ManifestSvc-->>CloneSvc: Success
    
    CloneSvc->>User: Offer to persist packageDir to config
    User-->>CloneSvc: Confirmation
    CloneSvc->>FS: Write config file
    
    CloneSvc-->>User: Clone complete
```

### Adopt Operation Flow

```mermaid
sequenceDiagram
    participant User
    participant AdoptSvc as AdoptService
    participant FS as Filesystem
    participant Executor as Executor
    participant ManifestSvc as ManifestService
    
    User->>AdoptSvc: Adopt([files], package)
    
    AdoptSvc->>AdoptSvc: PlanAdopt(files, package)
    
    loop For each file
        AdoptSvc->>AdoptSvc: resolveAdoptPath (absolute/relative/~)
        AdoptSvc->>FS: Check if exists
        AdoptSvc->>AdoptSvc: validateAdoptSource (check if already managed)
        AdoptSvc->>FS: IsDir?
        
        alt Is Directory
            AdoptSvc->>AdoptSvc: collectDirectoryFiles (recursive)
            
            rect rgb(80, 70, 100)
                note right of AdoptSvc: Directory Flattening
                AdoptSvc->>AdoptSvc: Create dir create operations
                AdoptSvc->>AdoptSvc: Create file move operations (translated names)
                AdoptSvc->>AdoptSvc: Create dir delete operations (deepest first)
                AdoptSvc->>AdoptSvc: Create link to package root
            end
        else Is File
            AdoptSvc->>AdoptSvc: UntranslateDotfile (.bashrc -> dot-bashrc)
            AdoptSvc->>AdoptSvc: Create move operation
            AdoptSvc->>AdoptSvc: Create link operation
        end
    end
    
    AdoptSvc->>Executor: Execute(plan)
    
    rect rgb(80, 100, 50)
        note right of Executor: Transactional Execution
        Executor->>Executor: Validate & checkpoint
        loop For each operation
            alt File move
                Executor->>FS: Rename or copy+delete (cross-device)
            else Dir create
                Executor->>FS: MkdirAll
            else Dir delete
                Executor->>FS: Remove
            else Link create
                Executor->>FS: Symlink
            end
        end
        
        alt Success
            Executor-->>AdoptSvc: Success
        else Failure
            Executor->>FS: Rollback all operations
            Executor-->>AdoptSvc: Error
        end
    end
    
    AdoptSvc->>ManifestSvc: UpdateWithSource(source=adopted)
    ManifestSvc->>ManifestSvc: Mark as "adopted" source
    ManifestSvc->>FS: Write manifest
```

## Type System

### Phantom Types for Path Safety

Phantom types encode path semantics at the type level:

```go
// PackagePath represents a path within the package directory
type PackagePath struct {
    path string
}

// TargetPath represents a path in the target directory
type TargetPath struct {
    path string
}

// FilePath represents a generic file path
type FilePath struct {
    path string
}
```

**Benefits**:
- Compile-time prevention of path mix-ups
- Self-documenting function signatures
- Type-guided refactoring
- Elimination of path-related bugs

**Usage Example**:
```go
// Function signature clearly indicates path expectations
func scanPackage(path PackagePath) Result[Package]

// Compiler prevents incorrect usage
scanPackage(targetPath)  // Compile error: type mismatch
```

### Result Type for Error Handling

The `Result[T]` type provides monadic error handling:

```go
type Result[T any] struct {
    value T
    err   error
    isOk  bool
}

func (r Result[T]) IsOk() bool
func (r Result[T]) IsErr() bool
func (r Result[T]) Unwrap() T
func (r Result[T]) UnwrapErr() error
func (r Result[T]) UnwrapOr(defaultValue T) T
func (r Result[T]) OrElse(fn func() T) T
func (r Result[T]) OrDefault() T
```

**Monadic Operations**:
```go
func Map[T, U any](r Result[T], fn func(T) U) Result[U]
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U]
func Collect[T any](results []Result[T]) Result[[]T]
```

**Benefits**:
- Explicit success or failure states
- Type-safe value extraction
- Composable error handling
- No nil pointer dereferencing
- Functional composition support

### Operation Types

Operations are represented as an interface with 9 concrete implementations:

```go
type Operation interface {
    ID() OperationID
    Kind() OperationKind
    Validate() error
    Dependencies() []Operation
    Execute(ctx context.Context, fs FS) error
    Rollback(ctx context.Context, fs FS) error
    String() string
    Equals(other Operation) bool
}

// Concrete operation types:
type LinkCreate struct { OpID OperationID; Source FilePath; Target TargetPath }
type LinkDelete struct { OpID OperationID; Target TargetPath }
type DirCreate struct { OpID OperationID; Path FilePath }
type DirDelete struct { OpID OperationID; Path FilePath }
type DirRemoveAll struct { OpID OperationID; Path FilePath }
type FileMove struct { OpID OperationID; Source TargetPath; Dest FilePath }
type FileBackup struct { OpID OperationID; Source FilePath; Backup FilePath }
type FileDelete struct { OpID OperationID; Path FilePath }
type DirCopy struct { OpID OperationID; Source FilePath; Dest FilePath }
```

**Operation Kinds**:
- `OpKindLinkCreate`: Create symbolic link
- `OpKindLinkDelete`: Remove symbolic link
- `OpKindDirCreate`: Create directory
- `OpKindDirDelete`: Remove empty directory
- `OpKindDirRemoveAll`: Recursively remove directory
- `OpKindFileMove`: Move file (handles cross-device moves via copy+delete)
- `OpKindFileBackup`: Create backup copy
- `OpKindFileDelete`: Delete file
- `OpKindDirCopy`: Recursive directory copy

## Error Handling

### Error Type Hierarchy

Domain-specific errors with rich context:

```go
// Core errors
type ErrInvalidPath struct { Path string; Reason string }
type ErrPackageNotFound struct { Package string }
type ErrConflict struct { Path string; Reason string }
type ErrSourceNotFound struct { Path string }
type ErrParentNotFound struct { Path string }
type ErrPermissionDenied struct { Path string; Operation string }

// Execution errors
type ErrExecutionFailed struct {
    Executed   int
    Failed     int
    RolledBack int
    Errors     []error
}

// Planning errors
type ErrCyclicDependency struct { Cycle []Operation }
type ErrEmptyPlan struct {}

// Service errors
type ErrMultiple struct { Errors []error }
type ErrAuthFailed struct { Cause error }
type ErrCloneFailed struct { URL string; Cause error }
type ErrInvalidBootstrap struct { Reason string; Cause error }
type ErrProfileNotFound struct { Profile string }
type ErrPackageDirNotEmpty struct { Path string; Cause error }
type ErrBootstrapExists struct { Path string }
```

### Error Wrapping

Errors are wrapped with context using `fmt.Errorf` and `%w`:

```go
if err := operation.Execute(); err != nil {
    return fmt.Errorf("failed to execute %s: %w", operation.Kind(), err)
}
```

### Error Aggregation

Multiple errors are collected and reported together:

```go
type ExecutionResult struct {
    Executed   []OperationID
    Failed     []OperationID
    Errors     []error
    RolledBack []OperationID
}
```

## Concurrency Model

### Thread Safety

All public API operations are safe for concurrent use:

```go
client, _ := dot.NewClient(config)

// Safe to call from multiple goroutines
go client.Manage(ctx, "vim")
go client.Status(ctx)
```

### Parallel Execution

The planner computes parallel execution batches:

1. **Dependency Analysis**: Build dependency graph
2. **Topological Sort**: Order operations respecting dependencies
3. **Batch Computation**: Group independent operations
4. **Parallel Execution**: Execute batches concurrently

**Example**:
```
Batch 1 (parallel):
  - CreateDir ~/.config
  - CreateDir ~/.local

Batch 2 (parallel, depends on Batch 1):
  - CreateLink ~/.config/nvim
  - CreateLink ~/.local/bin/script

Batch 3 (depends on Batch 2):
  - CreateLink ~/.config/nvim/init.vim
```

```mermaid
graph TB
    subgraph "Batch 1 - Parallel Execution"
        A[CreateDir<br/>~/.config]:::batch1
        B[CreateDir<br/>~/.local]:::batch1
        C[CreateDir<br/>~/.cache]:::batch1
    end
    
    subgraph "Batch 2 - Parallel Execution"
        D[CreateLink<br/>~/.config/nvim]:::batch2
        E[CreateLink<br/>~/.local/bin]:::batch2
        F[CreateLink<br/>~/.cache/app]:::batch2
    end
    
    subgraph "Batch 3 - Parallel Execution"
        G[CreateLink<br/>~/.config/nvim/init.vim]:::batch3
        H[CreateLink<br/>~/.config/nvim/lua]:::batch3
        I[CreateLink<br/>~/.local/bin/script]:::batch3
    end
    
    subgraph "Batch 4 - Sequential"
        J[CreateLink<br/>~/.config/nvim/lua/config.lua]:::batch4
    end
    
    A --> D
    A --> G
    A --> H
    B --> E
    B --> I
    C --> F
    
    D --> G
    D --> H
    E --> I
    
    G --> J
    H --> J
    
    style A fill:#3498DB,stroke:#1F618D,color:#fff
    style B fill:#3498DB,stroke:#1F618D,color:#fff
    style C fill:#3498DB,stroke:#1F618D,color:#fff
    style D fill:#50C878,stroke:#2D7A4A,color:#fff
    style E fill:#50C878,stroke:#2D7A4A,color:#fff
    style F fill:#50C878,stroke:#2D7A4A,color:#fff
    style G fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style H fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style I fill:#9B59B6,stroke:#6C3A7C,color:#fff
    style J fill:#E67E22,stroke:#A84E0F,color:#fff
    
    classDef batch1 stroke-width:3px
    classDef batch2 stroke-width:3px
    classDef batch3 stroke-width:3px
    classDef batch4 stroke-width:3px
```

### Doctor Parallel Scanning

The doctor service uses a worker pool for parallel directory scanning:

```go
// Determine worker count (default: NumCPU)
workers := scanCfg.MaxWorkers
if workers <= 0 {
    workers = runtime.NumCPU()
}

// Worker pool with cancellable context
workerCtx, cancel := context.WithCancel(ctx)
for i := 0; i < workers; i++ {
    go scanWorker(workerCtx, dirChan, resultChan)
}

// Early termination when MaxIssues reached
if len(issues) >= scanCfg.MaxIssues {
    cancel() // Cancel all workers
}
```

### Context Support

All operations support `context.Context` for cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := client.Manage(ctx, packages...)
// Respects context cancellation and timeout
```

## Testing Strategy

### Layer-Specific Testing

**Domain Layer**:
- Pure function testing
- Property-based testing of algebraic laws
- No filesystem access required
- Result type composition testing

**Core Layer**:
- Table-driven tests
- Edge case coverage
- In-memory filesystem for deterministic tests
- Ignore pattern matching tests

**Pipeline Layer**:
- Integration tests with memory filesystem
- Error propagation verification
- Context cancellation testing
- Targeted scanning optimization tests

**Executor Layer**:
- Rollback mechanism verification
- Checkpoint functionality
- Failure scenario coverage
- Parallel execution testing
- Precondition validation with pending operations

**API Layer**:
- End-to-end integration tests
- Service interaction testing
- Manifest persistence verification
- Hash-based change detection tests
- Backup restoration tests

**CLI Layer**:
- Command parsing tests
- Output format verification
- Error message validation
- Golden file testing for output consistency

### Test Coverage Requirements

- Minimum 80% code coverage
- Critical paths require 100% coverage
- All error paths must be tested
- Edge cases must have explicit tests

### Testing Tools

- Standard library `testing` package
- `testify/assert` for assertions
- Table-driven test pattern
- Memory-based filesystem adapter (billy/memfs)
- Golden file testing for outputs

## Dependency Rules

### Inward Dependencies

Dependencies flow inward toward the domain:

```mermaid
graph TD
    CLI[CLI Layer<br/>cmd/dot/]:::cliLayer
    API[API Layer<br/>pkg/dot/]:::apiLayer
    
    Pipeline[Pipeline Layer<br/>internal/pipeline/]:::middlewareLayer
    Executor[Executor Layer<br/>internal/executor/]:::middlewareLayer
    
    Scanner[Scanner<br/>internal/scanner/]:::coreLayer
    Planner[Planner<br/>internal/planner/]:::coreLayer
    Ignore[Ignore<br/>internal/ignore/]:::coreLayer
    
    Adapters[Adapters<br/>internal/adapters/]:::adapterLayer
    Manifest[Manifest<br/>internal/manifest/]:::adapterLayer
    Bootstrap[Bootstrap<br/>internal/bootstrap/]:::adapterLayer
    Config[Config<br/>internal/config/]:::adapterLayer
    Doctor[Doctor<br/>internal/doctor/]:::adapterLayer
    
    Domain[Domain Layer<br/>internal/domain/<br/><br/>No Dependencies<br/>Standard Library Only]:::domainLayer
    
    DomainPorts[Domain Ports<br/>Interfaces: FS, Logger,<br/>Tracer, Metrics]:::portsLayer
    
    CLI -->|depends on| API
    API -->|depends on| Pipeline
    API -->|depends on| Executor
    API -->|depends on| Scanner
    API -->|depends on| Planner
    API -->|depends on| Ignore
    API -->|depends on| Adapters
    API -->|depends on| Manifest
    API -->|depends on| Bootstrap
    API -->|depends on| Config
    
    Pipeline -->|depends on| Scanner
    Pipeline -->|depends on| Planner
    Pipeline -->|depends on| Ignore
    Pipeline -->|depends on| Domain
    Pipeline -->|depends on| Manifest
    
    Executor -->|depends on| Domain
    
    Scanner -->|depends on| Domain
    Planner -->|depends on| Domain
    Ignore -->|depends on| Domain
    
    Adapters -->|implements| DomainPorts
    Manifest -->|depends on| Domain
    Bootstrap -->|depends on| Domain
    Config -->|depends on| Domain
    Doctor -->|depends on| Domain
    
    DomainPorts -.defined in.-> Domain
    
    style CLI fill:#4A90E2,stroke:#2C5F8D,color:#fff,stroke-width:3px
    style API fill:#50C878,stroke:#2D7A4A,color:#fff,stroke-width:3px
    style Pipeline fill:#9B59B6,stroke:#6C3A7C,color:#fff,stroke-width:2px
    style Executor fill:#E67E22,stroke:#A84E0F,color:#fff,stroke-width:2px
    style Scanner fill:#3498DB,stroke:#1F618D,color:#fff,stroke-width:2px
    style Planner fill:#3498DB,stroke:#1F618D,color:#fff,stroke-width:2px
    style Ignore fill:#3498DB,stroke:#1F618D,color:#fff,stroke-width:2px
    style Adapters fill:#95A5A6,stroke:#5D6D7E,color:#fff,stroke-width:2px
    style Manifest fill:#95A5A6,stroke:#5D6D7E,color:#fff,stroke-width:2px
    style Bootstrap fill:#95A5A6,stroke:#5D6D7E,color:#fff,stroke-width:2px
    style Config fill:#95A5A6,stroke:#5D6D7E,color:#fff,stroke-width:2px
    style Doctor fill:#95A5A6,stroke:#5D6D7E,color:#fff,stroke-width:2px
    style Domain fill:#2ECC71,stroke:#1E8449,color:#fff,stroke-width:4px
    style DomainPorts fill:#7F8C8D,stroke:#5D6D7E,color:#fff,stroke-width:2px,stroke-dasharray: 5 5
    
    classDef cliLayer stroke-width:3px
    classDef apiLayer stroke-width:3px
    classDef middlewareLayer stroke-width:2px
    classDef coreLayer stroke-width:2px
    classDef domainLayer stroke-width:4px
    classDef adapterLayer stroke-width:2px,stroke-dasharray: 5 5
    classDef portsLayer stroke-width:2px,stroke-dasharray: 5 5
```

**Rules**:
1. Domain layer has no dependencies (except standard library)
2. Core layer depends only on domain
3. Pipeline and Executor depend on domain and core
4. Adapters implement domain ports and depend only on domain
5. API layer depends on all internal layers
6. CLI layer depends only on API layer (not internal packages)

### Import Restrictions

**Prohibited**:
- Internal packages importing from `pkg/dot` (would create cycle)
- CLI importing from `internal/*` directly
- Domain importing from infrastructure packages
- Core importing from adapters

**Required**:
- All internal packages import from `internal/domain` for types
- API layer re-exports domain types for public consumption
- Type aliases in `pkg/dot` for stable public API

### Adapter Independence

Adapters are swappable implementations:

```go
// Production
cfg := dot.Config{
    FS:     adapters.NewOSFilesystem(),
    Logger: adapters.NewSlogLogger(os.Stderr),
}

// Testing
cfg := dot.Config{
    FS:     adapters.NewMemFilesystem(),
    Logger: adapters.NewNoopLogger(),
}

// Dry-run
cfg := dot.Config{
    FS:     adapters.NewNoopFilesystem(),
    Logger: adapters.NewSlogLogger(os.Stderr),
}
```

## Performance Characteristics

### Time Complexity

**Scanning**: O(n) where n is number of files in packages
**Planning**: O(m + e) where m is operations and e is dependency edges
**Topological Sort**: O(m + e) using depth-first search
**Execution (sequential)**: O(m) where m is number of operations
**Execution (parallel)**: O(b) where b is number of batches
**Doctor Orphan Scan (parallel)**: O(d/w) where d is directories and w is workers

### Space Complexity

**Manifest Storage**: O(p × l) where p is packages and l is links per package
**Dependency Graph**: O(m + e) where m is operations and e is edges
**Checkpoint**: O(m) to store operation state
**Content Hashes**: O(p × f) where p is packages and f is files per package

### Optimizations

1. **Targeted Current State Scanning**: Only checks paths relevant to desired state (not full directory traversal)
2. **Directory Folding**: Reduce symlink count when entire directory owned by package
3. **Incremental Updates**: Use SHA256 content hashing to detect changed packages
4. **Parallel Execution**: Execute independent operations concurrently in batches
5. **Parallel Doctor Scanning**: Use worker pool (NumCPU workers) for directory scanning
6. **Lazy Loading**: Load manifests on demand
7. **Efficient Scanning**: Skip ignored directories early in traversal
8. **Early Termination**: Stop scanning when MaxIssues limit reached
9. **Cross-Device Move Optimization**: Try rename first, fallback to copy+delete

## Security Considerations

### Path Traversal Prevention

- All paths validated before use
- Phantom types prevent path confusion
- Relative paths resolved before operations
- Symlink targets validated
- Package directory boundary enforcement

### Safe Rollback

- Checkpoint created before operations
- Atomic rollback on failure
- No partial state on errors
- Reverse order rollback

### Manifest Integrity

- Manifest stored in target directory (user-controlled)
- SHA256 content hashing for change detection
- Validation before loading
- DoctorState for triage decisions

### Authentication Security

- SSH agent support with key-based authentication
- Token-based authentication via environment variables
- No credential storage in configuration
- Authentication resolution per-operation

### Adoption Safety

- Validates files not already managed
- Prevents circular adoption
- Checks package directory boundaries
- Path resolution with tilde expansion

### Error Information Disclosure

- Error messages avoid exposing sensitive paths where possible
- Detailed errors logged but sanitized for display
- Security-relevant errors handled specially

## Future Architecture Considerations

### Potential Enhancements

1. **Distributed Locking**: Support for network filesystem coordination
2. **Streaming Manifest Updates**: Avoid full manifest rewrite for large installations
3. **Plugin System**: External conflict resolution strategies and custom operations
4. **Remote Package Sources**: Support for fetching packages from URLs/registries
5. **Advanced Caching**: Cache scanning results and hashes for large repositories
6. **Doctor Pattern Learning**: Machine learning for better categorization
7. **Webhook Integration**: Post-installation hooks for automation
8. **Differential Backups**: Only backup changed files

### Backward Compatibility

The architecture supports evolution while maintaining compatibility:

- Public API in `pkg/dot` with type aliases
- Internal implementation can change freely
- Manifest versioning for format changes (currently version 1.0)
- Deprecation warnings for API changes
- Bootstrap configuration schema versioning

## References

### Related Documentation

- [User Guide](../user/index.md) - End-user documentation
- [Contributing Guide](../../CONTRIBUTING.md) - Development guidelines
- [Release Workflow](release-workflow.md) - Release process
- [Doctor System CUJs](doctor-system-cujs.md) - Doctor critical user journeys
- [Doctor UX Design](doctor-system-ux-design.md) - Doctor user experience design
- [Testing Guide](testing.md) - Testing standards and practices

### External Resources

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Functional Core, Imperative Shell](https://www.destroyallsoftware.com/screencasts/catalog/functional-core-imperative-shell)
- [Go Module Documentation](https://golang.org/ref/mod)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
- [Result Type Pattern](https://doc.rust-lang.org/std/result/)


