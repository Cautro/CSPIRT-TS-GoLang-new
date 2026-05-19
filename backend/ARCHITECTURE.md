# DDD/Clean Architecture Implementation

## Overview

This backend follows **Domain-Driven Design (DDD)** combined with **Clean Architecture** principles. The goal is to create a scalable, maintainable codebase that separates concerns into distinct layers and features.

## Architecture Layers

### 1. **Domain Layer** (`internal/<feature>/domain/`)

The **core business logic** of the application. Domain layer:
- Contains domain entities and value objects (no database or framework code)
- Defines repository interfaces (contracts)
- Contains domain-specific errors and validation rules
- Is independent of all other layers
- Should be 100% testable without external dependencies

**Example**: `internal/auth/domain/auth.go`

### 2. **Application Layer** (`internal/<feature>/application/`)

Orchestrates the domain layer and coordinates between layers. Contains:

**a) Use Cases/Services** (`services/`)
- Business operations that implement specific user actions
- Orchestrate domain entities and repositories
- Handle cross-cutting concerns (transactions, logging)
- Receive input DTOs, return output DTOs

**b) Data Transfer Objects** (`dto/`)
- Define request/response structures for APIs
- Are separate from domain entities
- Are converted to/from domain entities in use cases

**c) Ports** (`ports.go`)
- Define interfaces for dependencies needed by the application layer
- These are implemented by the infrastructure layer
- Enable dependency injection and loose coupling

### 3. **Infrastructure Layer** (`internal/<feature>/infrastructure/`)

Technical implementations and framework-specific code:

**a) HTTP Handlers** (`handlers/`)
- Parse HTTP requests and convert to DTOs
- Call application layer use cases
- Format responses and set HTTP headers
- No business logic here - just request/response translation

**b) Persistence** (`persistence/`)
- Implement repository interfaces from domain layer
- Database operations (SQL queries, transactions)
- Adapt external libraries (JWT, bcrypt) to domain needs

## Shared Layer

**Location**: `internal/shared/`

Contains cross-cutting concerns and shared types:

- **`domain/`**: Core interfaces used across all features
  - `Repository` - base interface for all repositories
  - `UnitOfWork` - transaction management
  - `Logger` - logging interface
  - `EventPublisher` - event-driven architecture support

- **`entity/`**: Shared domain entities
  - `User` - user domain entity used by multiple features
  - `SafeUser` - user without sensitive data

- **`errors/`**: Domain error types
  - `DomainError` - standardized error handling
  - Error codes for consistent error responses

## Dependency Injection Container

**Location**: `internal/container/container.go`

Centralized place where all dependencies are created and managed:
- Singletons for repositories, services, and adapters
- Factory methods for creating service instances
- Lifecycle management

## Feature Structure Template

```
internal/<feature>/
├── domain/
│   └── <feature>_domain.go      # Domain entities and repository interfaces
├── application/
│   ├── services/
│   │   └── <feature>_service.go # Use cases (orchestration logic)
│   ├── dto/
│   │   └── <feature>_dto.go     # Request/response DTOs
│   └── ports.go                 # Interfaces (dependencies)
└── infrastructure/
    ├── handlers/
    │   └── <feature>_handler.go # HTTP handlers
    └── persistence/
        └── <feature>_repository.go # Repository implementations
```

## Adding A New Feature

### Step 1: Define Domain (`domain/`)
- Create domain entities
- Define repository interfaces
- Define domain-specific errors

### Step 2: Create Application Layer (`application/`)
- Create DTOs for request/response
- Define ports (dependency interfaces)
- Implement use cases (business logic orchestration)

### Step 3: Implement Infrastructure (`infrastructure/`)
- Create HTTP handlers for routing
- Implement repositories from domain interfaces
- Create adapters for external libraries

### Step 4: Register in Container
- Add repository creation methods
- Expose in factory functions

### Step 5: Register Routes in Router
- Use container to inject dependencies
- Register HTTP endpoints

## Key Principles

### 1. **Separation of Concerns**
- Each layer has a single responsibility
- Domain layer contains ONLY business logic
- Infrastructure layer handles technical details
- Application layer orchestrates both

### 2. **Dependency Inversion**
- High-level modules don't depend on low-level modules
- Both depend on abstractions (interfaces)
- Pass dependencies as constructor parameters

### 3. **No Framework Code in Domain**
- Domain layer is framework-agnostic
- No imports from `gin`, `sql`, `jwt` in domain
- Makes domain testable and reusable

### 4. **DTOs for API Boundaries**
- Never expose domain entities directly in API responses
- Use DTOs to decouple API contracts from domain models
- Allows safe refactoring without breaking clients

### 5. **Error Handling**
- Use domain errors for business logic failures
- Map domain errors to HTTP status codes in handlers
- Consistent error response format

## Best Practices

### 1. Keep Use Cases Focused
Each use case represents one user action. Avoid bloated services.

### 2. Use Context for Cancellation
Pass `context.Context` through use cases for proper resource management.

### 3. Comment Business Decisions
Explain WHY code exists, not WHAT it does. Code should be self-documenting.

### 4. Test Each Layer
- Domain tests: pure business logic
- Application tests: use cases with mocked dependencies
- Infrastructure tests: database operations

## Scaling Benefits

This architecture provides:
- **Easy feature development**: Clear structure for new code
- **Easy testing**: Mock interfaces at each layer
- **Easy refactoring**: Changes don't cascade between layers
- **Easy optimization**: Optimize persistence without affecting business logic
- **Code reuse**: Domain logic usable in different contexts (API, CLI, gRPC)
- **Team scalability**: Multiple teams can work independently
7. Add focused tests near the layer being changed.

## Rules Of Thumb

- Handlers should not know database details.
- Services should not know Gin details.
- Storage should not import handlers.
- Feature packages should own their repository interfaces.
- Shared helpers belong in `internal/utils` only when they are truly generic.
- Prefer small comments that explain why code exists, not what every line does.
