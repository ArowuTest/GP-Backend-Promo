# Mynumba Don Win - Admin Portal Technical Documentation

**Version:** 1.0
**Date:** May 11, 2025

## 1. System Architecture

The Mynumba Don Win Admin Portal consists of two main components:

*   **Backend Service:** A Go (Golang) application built with the Gin Gonic framework. It handles business logic, API requests, database interactions, user authentication, and authorization.
    *   Repository: `GP-Backend-Promo`
    *   Language: Go (version 1.21.x)
    *   Framework: Gin Gonic
    *   Database: PostgreSQL
    *   Authentication: JWT (JSON Web Tokens)
*   **Frontend Application:** A React single-page application (SPA) built using Vite. It provides the user interface for administrators to interact with the system.
    *   Repository: `GP-ADMIN-PROMO`
    *   Language: TypeScript, React
    *   Build Tool: Vite

### Data Flow:

1.  Admin users interact with the React frontend in their browsers.
2.  The frontend makes API calls to the Go backend service.
3.  The backend service processes requests, interacts with the PostgreSQL database for data storage and retrieval (e.g., user accounts, prize structures, draw results, audit logs).
4.  For draw eligibility, the backend is designed to interface with PostHog, which acts as middleware, pre-filtering participant data (MSISDNs and points) into daily draw cohorts. The gaming engine (backend) pulls this data from PostHog.
5.  Winner notifications are intended to be sent via SMS (MTN SMS gateway or alternative).

## 2. Backend API Endpoints

The backend exposes a RESTful API. All protected endpoints require a JWT Bearer token in the `Authorization` header.

Base URL: `/api/v1`

### 2.1. Authentication (`/auth`)

*   **`POST /login`**
    *   Description: Authenticates an admin user.
    *   Request Body: `{ "username": "user@example.com", "password": "yourpassword" }`
    *   Response (Success 200 OK): `{ "message": "Login successful", "token": "jwt_token_string", "user": { "id": "uuid", "email": "...", "role": "SuperAdmin", ... } }`
    *   Response (Error): 400, 401, 500

### 2.2. User Management (`/admin/users` - SuperAdmin Role Required)

*   **`POST /`**
    *   Description: Creates a new admin user.
    *   Request Body: `models.AdminUserInput` (username, email, password, firstName, lastName, role)
    *   Response (Success 201 Created): `{ "message": "Admin user created successfully", "user": models.AdminUser }`
*   **`GET /`**
    *   Description: Lists all admin users.
    *   Response (Success 200 OK): `[{ models.AdminUser }]`
*   **`GET /:id`**
    *   Description: Retrieves a specific admin user by ID.
    *   Response (Success 200 OK): `models.AdminUser`
*   **`PUT /:id`**
    *   Description: Updates an admin user.
    *   Request Body: `models.AdminUserInput` (fields to update)
    *   Response (Success 200 OK): `{ "message": "Admin user updated successfully", "user": models.AdminUser }`
*   **`DELETE /:id`**
    *   Description: Deletes an admin user.
    *   Response (Success 200 OK): `{ "message": "Admin user deleted successfully" }`
*   **`PUT /:id/status`** (Conceptual - may need specific handler)
    *   Description: Updates a user's active status.

### 2.3. Prize Structure Management (`/admin/prize-structures` - SuperAdmin, Admin Roles)

*   **`POST /`**
    *   Description: Creates a new prize structure.
    *   Request Body: `models.PrizeStructureInput` (name, isActive, prizes: [{name, value, quantity}])
    *   Response (Success 201 Created): `models.PrizeStructure`
*   **`GET /`**
    *   Description: Lists all prize structures.
    *   Response (Success 200 OK): `[{ models.PrizeStructure }]`
*   **`GET /:id`**
    *   Description: Retrieves a specific prize structure.
    *   Response (Success 200 OK): `models.PrizeStructure`
*   **`PUT /:id`**
    *   Description: Updates a prize structure.
    *   Request Body: `models.PrizeStructureInput`
    *   Response (Success 200 OK): `models.PrizeStructure`
*   **`DELETE /:id`**
    *   Description: Deletes a prize structure.
    *   Response (Success 200 OK): `{ "message": "Prize structure deleted" }`
*   **`PUT /:id/activate`** (Conceptual - may need specific handler)
    *   Description: Activates/deactivates a prize structure.

### 2.4. Draw Management (`/admin/draws`)

*   **`POST /execute`** (SuperAdmin Role Required)
    *   Description: Executes a draw for a given date (date likely passed in request body or determined by backend logic based on PostHog cohorts).
    *   Request Body: `{ "drawDate": "YYYY-MM-DD" }` (Example)
    *   Response (Success 200 OK): Draw results (structure TBD, e.g., `{ "drawId": "uuid", "winners": [...], "runnerUps": [...] }`)
*   **`GET /`** (SuperAdmin, Admin, SeniorUser Roles)
    *   Description: Lists past draws.
    *   Response (Success 200 OK): `[{ models.Draw }]` (Draw model to include ID, date, winners, etc.)
*   **`GET /:id`** (SuperAdmin, Admin, SeniorUser Roles)
    *   Description: Retrieves details of a specific past draw.
    *   Response (Success 200 OK): `models.Draw`

### 2.5. Winner Reports (`/admin/reports/winners` - All Roles with Report Access)

*   **`GET /`**
    *   Description: Lists winners, potentially with filters (drawDate, prizeId).
    *   Query Params: `drawDate`, `prizeId` (Examples)
    *   Response (Success 200 OK): `[{ WinnerReportEntry }]` (Structure to include masked MSISDN, prize, date, etc.)

### 2.6. Data Upload Audit Logs (`/admin/audits/data-uploads` or `/admin/reports/all/data-uploads` - SuperAdmin, Admin, AllReportUser Roles)

*   **`GET /`**
    *   Description: Lists data upload audit entries.
    *   Response (Success 200 OK): `[{ models.DataUploadAuditEntry }]`

## 3. Database Schema

The backend uses PostgreSQL. GORM's AutoMigrate feature is used to manage the schema based on the models defined in `/internal/models/models.go`.

Key Models/Tables:

*   **`admin_users`**
    *   `id` (UUID, Primary Key)
    *   `username` (string, unique)
    *   `email` (string, unique)
    *   `password_hash` (string)
    *   `salt` (string)
    *   `first_name` (string)
    *   `last_name` (string)
    *   `role` (enum: SuperAdmin, Admin, SeniorUser, WinnerReportsUser, AllReportUser)
    *   `status` (enum: Active, Inactive, Suspended)
    *   `last_login_at` (timestamp with time zone, nullable)
    *   `failed_login_attempts` (integer)
    *   `created_at`, `updated_at`, `deleted_at` (timestamps)
*   **`prize_structures`**
    *   `id` (UUID, Primary Key)
    *   `name` (string)
    *   `is_active` (boolean)
    *   `created_at`, `updated_at`, `deleted_at`
*   **`prizes`** (associated with a prize structure)
    *   `id` (UUID, Primary Key)
    *   `prize_structure_id` (UUID, Foreign Key to `prize_structures`)
    *   `name` (string, e.g., "Jackpot", "Consolation")
    *   `value` (string, e.g., "N1,000,000", "N10,000 Airtime")
    *   `quantity` (integer, number of winners for this prize tier)
    *   `created_at`, `updated_at`, `deleted_at`
*   **`draws`** (records of executed draws)
    *   `id` (UUID, Primary Key)
    *   `draw_date` (date)
    *   `prize_structure_id` (UUID, Foreign Key to `prize_structures` used for this draw)
    *   `executed_by_user_id` (UUID, Foreign Key to `admin_users`)
    *   `status` (string, e.g., "Completed", "Failed")
    *   `eligible_participants_count` (integer)
    *   `total_points_in_draw` (integer)
    *   `created_at`, `updated_at`
*   **`draw_winners`** (winners for each prize in a draw)
    *   `id` (UUID, Primary Key)
    *   `draw_id` (UUID, Foreign Key to `draws`)
    *   `prize_id` (UUID, Foreign Key to `prizes`)
    *   `msisdn` (string, masked for display, full stored if necessary and compliant)
    *   `is_runner_up` (boolean)
    *   `runner_up_rank` (integer, if `is_runner_up` is true)
    *   `points_at_win` (integer, optional)
    *   `notification_status` (string, e.g., "Sent", "Failed", "Pending")
    *   `claim_status` (string, e.g., "Claimed", "Forfeited", "Pending")
    *   `created_at`, `updated_at`
*   **`data_upload_audit_entries`**
    *   `id` (UUID, Primary Key)
    *   `uploaded_by_user_id` (UUID, Foreign Key to `admin_users`)
    *   `upload_timestamp` (timestamp with time zone)
    *   `file_name` (string, optional)
    *   `record_count` (integer)
    *   `status` (string, e.g., "Success", "Partial Success", "Failure")
    *   `notes` (text, optional, for errors or details)
    *   `created_at`

## 4. Key Logic Flows

*   **User Authentication:** JWT generation on login, middleware validation on protected routes.
*   **Role-Based Access Control (RBAC):** Middleware in Gin (`auth.RoleAuthMiddleware`) checks user role from JWT claims against required roles for specific endpoints/groups. Frontend components also use the user role from `AuthContext` to conditionally render UI.
*   **Draw Execution:**
    1.  SuperAdmin selects a date via UI.
    2.  Frontend requests draw details for that date.
    3.  Backend determines eligible PostHog cohort based on date/day.
    4.  Backend fetches MSISDNs and points from PostHog (details of this interface TBD, assumed API call or direct DB query to PostHog data store).
    5.  Backend determines applicable prize structure.
    6.  UI displays pre-draw info (participant count, points, prize structure).
    7.  SuperAdmin clicks "Execute Draw".
    8.  Backend performs weighted random selection based on points (each point = one entry, MSISDN cannot win more than once per draw).
    9.  Winners and runner-ups are selected for each prize tier.
    10. Results are stored in `draws` and `draw_winners` tables.
    11. Results (masked) are returned to UI.
*   **Prize Structure Management:** Admins can define multiple prize tiers (name, value, quantity of winners) within a prize structure. These structures can be activated/deactivated.

## 5. Frontend Architecture

*   **React with Vite:** Modern, fast build tooling.
*   **TypeScript:** For type safety.
*   **React Router:** For client-side routing.
*   **Context API (`AuthContext`):** For managing global authentication state and user role.
*   **Component-Based Structure:** UI is broken down into reusable components (e.g., `UserListComponent`, `DrawExecutionPage`).
*   **Services (`authService`, etc.):** Abstract API calls to the backend.
*   **Styling:** (Assumed to be CSS Modules, Tailwind CSS, or a UI library like Material-UI/Ant Design - details depend on actual implementation choices in `GP-ADMIN-PROMO`).

## 6. Environment Configuration

*   **Backend:** Uses `.env` file for database credentials, JWT secret, port, etc. (See Backend Deployment Guide).
*   **Frontend:** Uses `.env.production` (and `.env.development`) for `VITE_API_BASE_URL`. (See Frontend Deployment Guide).

This document provides a high-level overview. For detailed code implementation, refer to the source code in the respective repositories.

