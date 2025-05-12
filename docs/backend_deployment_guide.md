# Backend Deployment Guide (Go - Gin Gonic)

This guide outlines the steps to build, configure, and deploy the Go backend application for the Mynumba Don Win Admin Portal.

## 1. Prerequisites

*   **Go:** Version 1.21.x installed on the deployment server. (Note: The sandbox environment currently has Go 1.18.1. For production, ensure Go 1.21 or the version specified in `go.mod` is used).
*   **PostgreSQL:** A running PostgreSQL database instance accessible from the deployment server.
*   **Environment Variables:** Access to set environment variables for configuration.

## 2. Source Code

*   Obtain the latest version of the backend source code from the repository (e.g., `/home/ubuntu/GP-Backend-Promo`).

## 3. Configuration

Create a `.env` file in the root of the backend project directory (`/home/ubuntu/GP-Backend-Promo/.env`) with the following environment variables:

```env
# Database Configuration
DB_HOST=your_db_host
DB_PORT=your_db_port
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
DB_SSLMODE=disable # or require, verify-full, etc., based on your DB setup

# JWT Secret Key (MUST be a strong, random string)
JWT_SECRET_KEY=your_very_strong_and_long_jwt_secret_key_at_least_32_bytes

# Server Port (Optional, defaults to 8080 if not set)
PORT=8080

# Gin Mode (Optional, defaults to debug, set to release for production)
GIN_MODE=release

# CORS Allowed Origins (Comma-separated list of frontend URLs)
# Example: CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com,http://localhost:3001
CORS_ALLOWED_ORIGINS=https://gp-admin-promo.vercel.app
```

**Important Security Note:** `JWT_SECRET_KEY` must be a cryptographically strong random string and kept confidential. Do not use default or weak keys in production.

## 4. Build the Application

Navigate to the backend project directory and build the executable:

```bash
cd /path/to/your/GP-Backend-Promo

# Ensure Go version matches go.mod (e.g., 1.21)
# If go.mod specifies a different version, update it or install the correct Go version.
# Example: go mod edit -go 1.21

# Tidy dependencies
go mod tidy

# Build the executable (e.g., named 'mynumba_backend')
go build -o mynumba_backend ./cmd/server/main.go
```

This will create an executable file named `mynumba_backend` (or your chosen name) in the project root.

## 5. Running the Application

Once built, you can run the application:

```bash
./mynumba_backend
```

The server will start, typically on the port specified by the `PORT` environment variable (default 8080).

## 6. Deployment as a Service (Recommended for Production)

For production environments, it is highly recommended to run the backend application as a system service (e.g., using systemd on Linux) to ensure it runs continuously and restarts automatically on failure.

### Example systemd Service File (`mynumba-backend.service`):

Create a file `/etc/systemd/system/mynumba-backend.service` with content similar to the following (adjust paths and user as needed):

```ini
[Unit]
Description=Mynumba Don Win Backend Service
After=network.target postgresql.service # Ensure network and DB are up

[Service]
User=your_deploy_user # The user the service will run as
Group=your_deploy_group
WorkingDirectory=/path/to/your/GP-Backend-Promo # Path to the application
EnvironmentFile=/path/to/your/GP-Backend-Promo/.env # Path to your .env file
ExecStart=/path/to/your/GP-Backend-Promo/mynumba_backend # Command to start the application
Restart=always
RestartSec=10s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=mynumba-backend

[Install]
WantedBy=multi-user.target
```

**Steps to enable and start the service:**

1.  **Reload systemd:** `sudo systemctl daemon-reload`
2.  **Enable the service to start on boot:** `sudo systemctl enable mynumba-backend.service`
3.  **Start the service immediately:** `sudo systemctl start mynumba-backend.service`
4.  **Check the service status:** `sudo systemctl status mynumba-backend.service`
5.  **View logs:** `sudo journalctl -u mynumba-backend -f`

## 7. Reverse Proxy (Recommended)

It is also recommended to use a reverse proxy like Nginx or Apache in front of the Go application. This can handle SSL termination, load balancing (if needed), serving static files (if any, though not typical for this backend), and provide an additional layer of security.

### Example Nginx Configuration Snippet:

```nginx
server {
    listen 80;
    server_name your_backend_api_domain.com;

    # Redirect HTTP to HTTPS (if SSL is configured)
    # return 301 https://$host$request_uri;

    location / {
        proxy_pass http://localhost:8080; # Assuming Go app runs on port 8080
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# If using SSL (recommended):
# server {
#     listen 443 ssl http2;
#     server_name your_backend_api_domain.com;
# 
#     ssl_certificate /path/to/your/fullchain.pem;
#     ssl_certificate_key /path/to/your/privkey.pem;
# 
#     # ... other SSL configurations ...
# 
#     location / {
#         proxy_pass http://localhost:8080;
#         proxy_set_header Host $host;
#         proxy_set_header X-Real-IP $remote_addr;
#         proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
#         proxy_set_header X-Forwarded-Proto $scheme;
#     }
# }
```

## 8. Database Migrations

The application uses GORM's AutoMigrate feature (`config.DB.AutoMigrate(&models.AdminUser{}, ...)` in `internal/config/db.go`) to automatically create or update database tables based on the defined models. This typically runs when the application starts.

For production, consider more robust migration strategies if complex schema changes are anticipated, but for the current setup, AutoMigrate handles basic schema evolution.

Ensure the database user specified in `.env` has permissions to create and alter tables in the specified database.

