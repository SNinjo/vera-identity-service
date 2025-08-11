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
| `repository.go` | Define data structures and database access for the feature |
| `dto.go` | Define data transfer objects for API contracts |

### 3.3 Naming Conventions

Naming follows the [Google Go Style Guide](https://google.github.io/styleguide/go/decisions.html#naming):

- **Package names** should be short, all lowercase, and without underscores (e.g., `featurea`, not `featureA` or `feature_a`).
- **Directories and files** are named after features (e.g., `user`), using lowercase and no special characters.
- **File names** should reflect their role: `routes.go`, `handler.go`, `service.go`, `repository.go`, `dto.go`.
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

#### 6.1.1 Test Scenarios

**Happy Path Scenarios:**

- Test normal operation with valid inputs
- Verify expected outputs and side effects
- Test with typical data ranges and formats

**Edge Cases:**

- Test boundary conditions (empty strings, null values, max/min values)
- Test with unusual but valid inputs
- Test performance with large datasets

**Error Conditions:**

- Test with invalid inputs (malformed data, wrong types)
- Test error handling and recovery
- Verify appropriate error messages and codes

#### 6.1.2 Test Structure & Best Practices

- Each test targets a single function, method, or class
- Mock or stub dependencies to isolate the code under test
- Place unit tests in the same package as the code, using the `_test.go` suffix
- Use the Arrange-Act-Assert (AAA) pattern for consistent test structure
- Use assertion libraries like `testify/assert` for clarity
- Name tests following the format: `Test<role>_<method>_<scenario>` (e.g., `TestService_Add_Success`)
- Fast execution (< 100ms per test) with no external dependencies
- Use table-driven tests for multiple input scenarios

### 6.2 API Testing

API tests verify that HTTP endpoints behave as expected for both valid and invalid requests.

#### 6.2.1 Test Focus

Primary Focus: Happy Path Scenarios

- Test complete user journeys and workflows
- Verify correct request/response formats
- Test authentication and authorization flows
- Validate business logic integration

Secondary Focus: Critical Error Cases

- Test common error scenarios (validation failures, not found)
- Verify error response formats and codes
- Test rate limiting and security measures

#### 6.2.2 Organization & Naming

- Group tests by endpoint, usually in their own test file
- Each test covers a specific scenario (success, failure, edge case)
- Name tests following the format: `TestAPI_<name>_<scenario>` (e.g., `TestAPI_GetUser_Success`)
- Use descriptive test names that explain the business scenario

#### 6.2.3 Assertions & Best Practices

- Always check the HTTP status code
- For error responses, assert the error code (HTTP status and/or `code` field in the response body)
- Validate the response body structure and content when relevant
- Use the Arrange-Act-Assert (AAA) pattern for consistent test structure
- Use assertion libraries like `testify/assert` for clarity
- Test response headers when relevant (content-type, authorization, etc.)

#### 6.2.4 Mocking & External Services

- Use tools like testcontainers-go to run real or mock services in containers
- Set up test databases and seed data to verify authentication, CRUD operations, and error scenarios
- Clean up containers or mocks after tests
- Use in-memory databases for faster test execution when possible

### 6.4 Test Coverage Requirements

#### 6.4.1 Unit Tests

- **Minimum coverage**: 80% of business logic code
- **Critical paths**: 100% coverage for authentication, authorization, and data validation
- **Focus areas**: Service layer, utility functions, data transformations

#### 6.4.2 API Tests

- **Coverage**: All public endpoints with happy path and critical error scenarios
- **Authentication**: All protected endpoints with valid and invalid tokens
- **Validation**: All input validation rules and error responses

### 6.5 Continuous Integration

#### 6.5.1 Test Execution

- Run unit tests on every commit (fast feedback)
- Run API tests on pull requests (medium feedback)
- Fail builds on test failures with clear error reporting

#### 6.5.2 Test Environment

- Maintain dedicated test environments for different test types
- Use containerization for consistent test environments
- Implement proper cleanup and resource management
- Monitor test execution time and optimize slow tests

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
