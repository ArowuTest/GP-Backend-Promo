# Repository Structure Diagram

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

# Clean Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ┌─────────────────────────┐                                    │
│  │                         │                                    │
│  │      Domain Layer       │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  Draw Domain    │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │ Participant     │    │                                    │
│  │  │ Domain          │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  Prize Domain   │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  Audit Domain   │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  User Domain    │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  └─────────────────────────┘                                    │
│                                                                 │
│  ┌─────────────────────────┐                                    │
│  │                         │                                    │
│  │   Application Layer     │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  Draw Use Cases │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  Participant    │    │                                    │
│  │  │  Use Cases      │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  Prize Use Cases│    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  Audit Use Cases│    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  │  ┌─────────────────┐    │                                    │
│  │  │  User Use Cases │    │                                    │
│  │  └─────────────────┘    │                                    │
│  │                         │                                    │
│  └─────────────────────────┘                                    │
│                                                                 │
│  ┌─────────────────────────┐        ┌─────────────────────────┐ │
│  │                         │        │                         │ │
│  │  Infrastructure Layer   │        │    Interface Layer      │ │
│  │                         │        │                         │ │
│  │  ┌─────────────────┐    │        │  ┌─────────────────┐    │ │
│  │  │  GORM           │    │        │  │  API Handlers   │    │ │
│  │  │  Repositories   │    │        │  └─────────────────┘    │ │
│  │  └─────────────────┘    │        │                         │ │
│  │                         │        │  ┌─────────────────┐    │ │
│  │  ┌─────────────────┐    │        │  │  DTOs           │    │ │
│  │  │  Configuration  │    │        │  └─────────────────┘    │ │
│  │  └─────────────────┘    │        │                         │ │
│  │                         │        │  ┌─────────────────┐    │ │
│  │  ┌─────────────────┐    │        │  │  Middleware     │    │ │
│  │  │  External       │    │        │  └─────────────────┘    │ │
│  │  │  Integrations   │    │        │                         │ │
│  │  └─────────────────┘    │        │  ┌─────────────────┐    │ │
│  │                         │        │  │  Router         │    │ │
│  └─────────────────────────┘        │  └─────────────────┘    │ │
│                                     │                         │ │
│                                     └─────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

# Dependency Flow Diagram

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│               │     │               │     │               │     │               │
│  Interface    │     │  Application  │     │    Domain     │     │Infrastructure │
│    Layer      │ ──> │    Layer      │ ──> │    Layer      │ <── │    Layer      │
│               │     │               │     │               │     │               │
└───────────────┘     └───────────────┘     └───────────────┘     └───────────────┘
```

# API Endpoints Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ┌─────────────────────────┐                                    │
│  │                         │                                    │
│  │      Draw API           │                                    │
│  │                         │                                    │
│  │  GET  /api/v1/admin/draws/eligibility-stats                 │
│  │  POST /api/v1/admin/draws/execute                           │
│  │  POST /api/v1/admin/draws/invoke-runner-up                  │
│  │  GET  /api/v1/admin/draws                                   │
│  │  GET  /api/v1/admin/draws/:id                               │
│  │  GET  /api/v1/admin/winners                                 │
│  │  PUT  /api/v1/admin/winners/:id/payment-status              │
│  │                         │                                    │
│  └─────────────────────────┘                                    │
│                                                                 │
│  ┌─────────────────────────┐                                    │
│  │                         │                                    │
│  │    Prize API            │                                    │
│  │                         │                                    │
│  │  GET    /api/v1/admin/prize-structures                      │
│  │  POST   /api/v1/admin/prize-structures                      │
│  │  GET    /api/v1/admin/prize-structures/:id                  │
│  │  PUT    /api/v1/admin/prize-structures/:id                  │
│  │  DELETE /api/v1/admin/prize-structures/:id                  │
│  │                         │                                    │
│  └─────────────────────────┘                                    │
│                                                                 │
│  ┌─────────────────────────┐                                    │
│  │                         │                                    │
│  │   Participant API       │                                    │
│  │                         │                                    │
│  │  POST   /api/v1/admin/participants/upload                   │
│  │  GET    /api/v1/admin/participants/stats                    │
│  │  GET    /api/v1/admin/participants/uploads                  │
│  │  GET    /api/v1/admin/participants                          │
│  │  DELETE /api/v1/admin/participants/uploads/:id              │
│  │                         │                                    │
│  └─────────────────────────┘                                    │
│                                                                 │
│  ┌─────────────────────────┐                                    │
│  │                         │                                    │
│  │     Audit API           │                                    │
│  │                         │                                    │
│  │  GET    /api/v1/admin/reports/data-uploads                  │
│  │                         │                                    │
│  └─────────────────────────┘                                    │
│                                                                 │
│  ┌─────────────────────────┐                                    │
│  │                         │                                    │
│  │     User API            │                                    │
│  │                         │                                    │
│  │  POST   /api/v1/admin/auth/login                            │
│  │  GET    /api/v1/admin/users                                 │
│  │  POST   /api/v1/admin/users                                 │
│  │  GET    /api/v1/admin/users/:id                             │
│  │  PUT    /api/v1/admin/users/:id                             │
│  │                         │                                    │
│  └─────────────────────────┘                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```
