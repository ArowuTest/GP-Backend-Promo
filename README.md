# GP-Backend-Promo

Backend service for the Bridgetunes 'MyNumba Don Win' promotion, implementing a clean architecture pattern.

## Architecture Overview

This project follows the clean architecture pattern with the following layers:

1. **Domain Layer** - Core business logic and entities
2. **Application Layer** - Use cases and application services
3. **Infrastructure Layer** - External dependencies and implementations
4. **Interface Layer** - API handlers, DTOs, and middleware

## Key Features

- Draw execution and management
- Prize structure configuration
- Participant data management
- Winner selection and runner-up handling
- Audit logging and reporting
- User authentication and authorization

## Directory Structure

```
├── cmd
│   └── server          # Main application entry point
├── internal
│   ├── application     # Application services and use cases
│   ├── domain          # Domain entities and interfaces
│   ├── infrastructure  # External dependencies and implementations
│   └── interface       # API handlers, DTOs, and middleware
```

## Getting Started

### Prerequisites

- Go 1.19 or higher
- PostgreSQL database

### Installation

1. Clone the repository
2. Configure the environment variables
3. Run the application

```bash
go run cmd/server/main.go
```

## API Endpoints

The API follows a RESTful design with the following main endpoints:

- `/api/v1/auth/login` - User authentication
- `/api/v1/admin/draws` - Draw management
- `/api/v1/admin/prize-structures` - Prize structure management
- `/api/v1/admin/participants` - Participant data management
- `/api/v1/admin/winners` - Winner management
- `/api/v1/admin/users` - User management

## Development Notes

- All domain entities are properly isolated with their own interfaces
- Repository implementations use GORM for database access
- API handlers use Gin for routing and middleware
- Authentication uses JWT tokens
- Error handling is consistent across all layers

## Recent Improvements

- Refactored to clean architecture pattern
- Fixed package naming inconsistencies
- Aligned all DTOs with frontend contracts
- Implemented proper error handling and audit logging
- Added comprehensive repository implementations
