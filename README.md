# GP-Backend-Promo Clean Architecture Implementation

## Overview

This repository contains the clean architecture implementation of the GP-Backend-Promo backend for the Mynumba Don Win promotion. The implementation follows domain-driven design principles and clean architecture patterns to create a maintainable, testable, and scalable codebase.

## Repository Structure

```
GP-Backend-Promo/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point with DI setup
├── internal/
│   ├── domain/                     # Domain layer - core business entities
│   │   ├── audit/
│   │   │   └── entity.go           # Audit domain entities and interfaces
│   │   ├── draw/
│   │   │   └── entity.go           # Draw domain entities and interfaces
│   │   ├── participant/
│   │   │   └── entity.go           # Participant domain entities and interfaces
│   │   ├── prize/
│   │   │   └── entity.go           # Prize domain entities and interfaces
│   │   └── user/
│   │       └── entity.go           # User domain entities and interfaces
│   ├── application/                # Application layer - use cases
│   │   ├── audit/
│   │   │   ├── log_audit.go
│   │   │   └── get_audit_logs.go
│   │   ├── draw/
│   │   │   ├── execute_draw.go
│   │   │   ├── get_draw_details.go
│   │   │   ├── list_draws.go
│   │   │   ├── get_eligibility_stats.go
│   │   │   └── invoke_runner_up.go
│   │   ├── participant/
│   │   │   ├── upload_participants.go
│   │   │   └── get_participant_stats.go
│   │   ├── prize/
│   │   │   ├── create_prize_structure.go
│   │   │   ├── get_prize_structure.go
│   │   │   ├── list_prize_structures.go
│   │   │   └── update_prize_structure.go
│   │   └── user/
│   │       ├── authenticate_user.go
│   │       ├── create_user.go
│   │       ├── update_user.go
│   │       ├── get_user.go
│   │       └── list_users.go
│   ├── infrastructure/             # Infrastructure layer - external dependencies
│   │   ├── config/
│   │   │   ├── config.go           # Configuration management
│   │   │   └── db.go               # Database connection
│   │   └── persistence/
│   │       └── gorm/
│   │           ├── draw_repository.go
│   │           ├── participant_repository.go
│   │           ├── prize_repository.go
│   │           ├── audit_repository.go
│   │           └── user_repository.go
│   └── interface/                  # Interface layer - API endpoints
│       ├── api/
│       │   ├── router.go           # API routes setup
│       │   ├── handler/
│       │   │   ├── draw_handler.go
│       │   │   ├── participant_handler.go
│       │   │   ├── prize_handler.go
│       │   │   ├── audit_handler.go
│       │   │   └── user_handler.go
│       │   └── middleware/
│       │       ├── auth_middleware.go
│       │       ├── cors_middleware.go
│       │       └── error_middleware.go
│       └── dto/
│           ├── request/
│           │   └── request.go      # API request DTOs
│           └── response/
│               └── response.go     # API response DTOs
├── docs/                           # Documentation
│   ├── implementation_guide.md
│   ├── deployment_guide.md
│   ├── frontend_backend_alignment_validation.md
│   ├── cross_dependency_validation.md
│   └── repository_structure.md
├── Dockerfile                      # Container definition
├── .env.example                    # Environment variables template
├── go.mod                          # Go module definition
└── go.sum                          # Go module checksums
```

## Architecture Layers

### 1. Domain Layer

The domain layer contains the core business entities, repository interfaces, and domain-specific errors. This layer is independent of any external concerns and contains the business rules.

Key components:
- Domain entities (Draw, Participant, Prize, Audit, User)
- Repository interfaces
- Domain-specific error types and validation logic

### 2. Application Layer

The application layer contains the use cases that orchestrate the flow of data to and from the domain entities and implement the business rules.

Key components:
- Use cases for all domains
- Input/output DTOs for use cases
- Business logic implementation

### 3. Infrastructure Layer

The infrastructure layer contains implementations of the repository interfaces defined in the domain layer, as well as external service integrations.

Key components:
- GORM repository implementations
- Configuration management
- External service integrations (PostHog, SMS Gateway, MTN API)

### 4. Interface Layer

The interface layer contains the API handlers, middleware, and DTOs that handle HTTP requests and responses.

Key components:
- API handlers for all domains
- Request/response DTOs
- Middleware (Authentication, CORS, Error handling)
- Router configuration

## API Endpoints

### Draw Management
- `GET /api/v1/admin/draws/eligibility-stats` - Get eligibility statistics for a draw date
- `POST /api/v1/admin/draws/execute` - Execute a draw
- `POST /api/v1/admin/draws/invoke-runner-up` - Invoke a runner-up for a prize
- `GET /api/v1/admin/draws` - List all draws
- `GET /api/v1/admin/draws/:id` - Get details of a single draw
- `GET /api/v1/admin/winners` - List all winners
- `PUT /api/v1/admin/winners/:id/payment-status` - Update winner payment status

### Prize Structure Management
- `GET /api/v1/admin/prize-structures` - List all prize structures
- `POST /api/v1/admin/prize-structures` - Create a new prize structure
- `GET /api/v1/admin/prize-structures/:id` - Get a prize structure by ID
- `PUT /api/v1/admin/prize-structures/:id` - Update a prize structure
- `DELETE /api/v1/admin/prize-structures/:id` - Delete a prize structure

### Participant Management
- `POST /api/v1/admin/participants/upload` - Upload participant data CSV
- `GET /api/v1/admin/participants/stats` - Get participant statistics
- `GET /api/v1/admin/participants/uploads` - List participant upload audit records
- `GET /api/v1/admin/participants` - Get participants for a specific draw date
- `DELETE /api/v1/admin/participants/uploads/:id` - Delete a participant upload

### Audit Management
- `GET /api/v1/admin/reports/data-uploads` - Get data upload audit records

### User Management
- `POST /api/v1/admin/auth/login` - Authenticate user
- `GET /api/v1/admin/users` - List all users
- `POST /api/v1/admin/users` - Create a new user
- `GET /api/v1/admin/users/:id` - Get a user by ID
- `PUT /api/v1/admin/users/:id` - Update a user

## Getting Started

### Prerequisites

- Go 1.19 or higher
- PostgreSQL database

### Installation

1. Clone the repository:
```bash
git clone https://github.com/ArowuTest/GP-Backend-Promo.git
cd GP-Backend-Promo
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
   - Copy `.env.example` to `.env`
   - Update the values in `.env` with your configuration

4. Build and run the application:
```bash
go run cmd/server/main.go
```

### Deployment

For deployment instructions, please refer to the [Deployment Guide](docs/deployment_guide.md).

## Documentation

For more detailed information, please refer to the following documentation:

- [Implementation Guide](docs/implementation_guide.md)
- [Deployment Guide](docs/deployment_guide.md)
- [Frontend-Backend Alignment Validation](docs/frontend_backend_alignment_validation.md)
- [Cross-Dependency Validation](docs/cross_dependency_validation.md)
- [Repository Structure](docs/repository_structure.md)

## License

This project is proprietary and confidential.
