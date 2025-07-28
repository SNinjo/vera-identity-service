# Technology Specification

## 1. Overview

The Vera Identity Service is a Go-based microservice that provides authentication and authorization for the Vera ecosystem. It is designed with a feature-first architecture to maximize modularity, maintainability, and scalability.

## 2. Technology Stack

- Go v1.23.0
- Gin (HTTP web framework)
- GORM (ORM for PostgreSQL)
- Swaggo (Swagger/OpenAPI documentation)
- Google OAuth 2.0 (authentication provider)
- JWT (token-based authentication)

## 3. Project Structure

### 3.1 Directory Layout

The codebase is organized by feature/domain, following the spirit of the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) repository, but adapted for a feature-first approach.

```text
project-root/
├── cmd/                # Main applications for this project
│   └── myapp/
│       └── main.go
├── internal/           # Private application and library code
│   ├── featurea/       # Example feature module
│   ├── featureb/
│   └── ...
├── pkg/                # Public libraries for other projects
├── api/                # API definitions (OpenAPI/Protobuf/etc.)
├── configs/            # Configuration files
├── scripts/            # Utility scripts
├── docs/               # Documentation
├── test/               # Additional external test apps and test data
├── go.mod
├── go.sum
└── README.md
```

### 3.2 Module/Feature Layout

Each feature module follows a consistent structure:

| File | Purpose |
|------|---------|
| `routes.go` | Define HTTP endpoints and middleware for the feature |
| `handler.go` | Process HTTP requests and responses |
| `service.go` | Contain business logic and orchestration for the feature |
| `model.go` | Define data structures and database access for the feature |
| `dto.go` | Define data transfer objects for API contracts |

### 3.3 Naming Conventions

Naming follows the [Google Go Style Guide](https://google.github.io/styleguide/go/decisions.html#naming):

- **Package names** should be short, all lowercase, and without underscores (e.g., `featurea`, not `featureA` or `feature_a`).
- **Directories and files** are named after features (e.g., `user`), using lowercase and no special characters.
- **File names** should reflect their role: `routes.go`, `handler.go`, `service.go`, `model.go`, `dto.go`.
- **Avoid** generic names like `util`, `common`, or `helper`.
- **Test files** use the `_test.go` suffix.
- **Initialisms** (like `ID`, `API`) should be capitalized (e.g., `userID`, not `userId`).

## 4. Configuration & Environment

### 4.1 Environment Variables

- Store sensitive configuration (DB credentials, secrets, OAuth keys) in environment variables.
- Use `.env` files for local development.

### 4.2 Configuration Files

- Centralized configuration in the `configs/` directory.
- Support for different environments (development, production).

## 5. Error Handling

### 5.1 Error Code Structure

Error codes follow the format: `<http_status>_<service_id>_<serial_number>`

**Components:**

- HTTP Status: The HTTP status code (400, 401, 404, 500, etc.)
- Service ID: Unique identifier for the service
- Serial Number: Unique identifier for the specific error type within the service

**Key Principles:**

- The same error type across different APIs within the same service uses the same error code
- Serial numbers are unique across ALL error codes within the service (no duplicates regardless of HTTP status)
- Serial numbers are assigned sequentially starting from 001
- Error codes represent error semantics, not location

**Examples:**

- `400_01_001` - Bad Request, Service 01, "Invalid email format" error
- `400_01_002` - Bad Request, Service 01, "Password too short" error  
- `401_01_003` - Unauthorized, Service 01, "Invalid credentials" error
- `404_01_004` - Not Found, Service 01, "User not found" error
- `500_01_005` - Internal Server Error, Service 01, "Database connection failed" error

### 5.2 Response Format

**Success Response (HTTP 200):**

```json
{
  "id": 1,
  "email": "user@example.com"
}
```

**Error Response (HTTP 4xx/5xx):**

```json
{
  "code": "400_01_001",
  "message": "Invalid email format",
  "timestamp": "1970-01-01T00:00:00Z"
}
```

### 5.3 Error Package Structure

Error codes are centralized in the `internal/apperror/` package:

```text
internal/apperror/
├── codes.go      # Error code constants (400_01_001, etc.)
└── ...
```

All features import error codes from this centralized location to ensure consistency across the application.

## 6. Testing

### 6.1 Unit Testing

Unit tests check that individual functions or methods work as intended, in isolation from the rest of the system.

- Each test targets a single function, method, or class.
- Mock or stub dependencies to isolate the code under test.
- Place unit tests in the same package as the code, using the `_test.go` suffix.
- Use clear, descriptive test names.
- Use assertion libraries like `testify/assert` for clarity.
- Name tests following the format: `TestUnit_<name>_<scenario>` (e.g., `TestUnit_Add_Success`).

### 6.2 API Testing

API tests check that HTTP endpoints behave as expected for both valid and invalid requests.

#### 6.2.1 Organization & Naming

- Group tests by endpoint, usually in their own test file.
- Each test covers a specific scenario (success, failure, edge case).
- Name tests following the format: `TestAPI_<name>_<scenario>` (e.g., `TestAPI_Login_Success`).

#### 6.2.2 Assertions & Best Practices

- Always check the HTTP status code.
- For error responses, assert the error code (HTTP status and/or `code` field in the response body). Checking the error message is optional.
- Validate the response body when relevant.
- Use a consistent structure: setup, execute, assert.
- Use assertion libraries like `testify/assert` for clarity.

#### 6.2.3 Mocking & External Services

- Use tools like testcontainers-go to run real or mock services in containers for tests that need external dependencies.
- Set up test databases and seed data to verify authentication, CRUD operations, and error scenarios.
- Clean up containers or mocks after tests.

## 7. Deployment & Operations

### 7.1 Build & Run

- Use Makefile for build and deployment scripts.
- Containerized deployment with Docker.

### 7.2 Monitoring

- Structured logging for traceability.
- Health check endpoints for service monitoring.

### 7.3 Scaling

- Stateless service design for horizontal scaling.
- Database connection pooling for performance.
