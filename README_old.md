# TestContainers Demo - Integration Testing with PostgreSQL

This project demonstrates integration testing using TestContainers with PostgreSQL in Go. It implements a complete CRUD (Create, Read, Update, Delete) user management system with comprehensive integration tests.

## Project Structure

```
testcontainers-demo/
├── go.mod                          # Go module definition
├── go.sum                          # Go module checksums
├── models/
│   └── user.go                     # User data model
├── repository/
│   ├── user_repository.go          # Database access layer
│   └── user_repository_test.go     # Integration tests
├── migrations/
│   └── init.sql                    # Database schema and test data
└── README.md                       # This file
```

## Features

### User Repository Operations

- **GetByID**: Retrieve user by ID
- **GetByEmail**: Retrieve user by email
- **Create**: Create new user
- **Update**: Update existing user
- **Delete**: Delete user
- **List**: List all users
- **FindByNamePattern**: Find users by name pattern
- **CountUsers**: Count total users
- **GetRecentUsers**: Get users created in last N days

### Integration Tests

- Complete CRUD operation testing
- Error case testing (non-existent users, duplicates)
- Pattern matching tests
- Transaction rollback testing
- Table-driven testing examples
- Test cleanup patterns

## Prerequisites

- Go 1.19 or higher
- Docker Desktop installed and running
- Basic understanding of Go and SQL

## Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd testcontainers-demo
```

2. Install dependencies:

```bash
go mod tidy
```

3. Ensure Docker is running:

```bash
docker ps
```

## Running Tests

### Run All Tests

```bash
go test ./repository -v
```

### Run Tests with Coverage

```bash
go test ./repository -v -cover
```

### Run Specific Test

```bash
go test ./repository -v -run TestGetByID
```

### Run Tests with Detailed Output

```bash
go test ./repository -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## How It Works

### TestContainers Lifecycle

1. **TestMain Setup**:

   - Starts PostgreSQL container with `postgres:15-alpine` image
   - Runs initialization scripts (`init.sql`)
   - Establishes database connection
   - Waits for database to be ready

2. **Test Execution**:

   - Each test runs against the real PostgreSQL database
   - Tests are isolated through cleanup mechanisms
   - Database state is managed between tests

3. **Cleanup**:
   - Container is automatically terminated after all tests
   - Database connection is closed
   - All test data is destroyed

### Test Strategies

- **Test Isolation**: Each test cleans up its data using `defer` statements
- **Table-Driven Tests**: Multiple test cases in single test function
- **Transaction Testing**: Verify rollback behavior
- **Error Testing**: Ensure proper error handling
- **Pattern Testing**: Test SQL pattern matching

## Key Components

### Database Schema

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Test Configuration

- **Database**: PostgreSQL 15 Alpine
- **Wait Strategy**: Wait for "database system is ready" log
- **Initialization**: Automatic schema creation and test data insertion
- **Cleanup**: Automatic container termination

## Running Individual Exercises

### Exercise 1: Basic Setup

```bash
go test ./repository -v -run "TestGetByID|TestGetByEmail"
```

### Exercise 2: CRUD Operations

```bash
go test ./repository -v -run "TestCreate|TestUpdate|TestDelete|TestList"
```

### Exercise 3: Advanced Queries

```bash
go test ./repository -v -run "TestFindByNamePattern|TestCountUsers|TestGetRecentUsers"
```

### Exercise 4: Transaction Testing

```bash
go test ./repository -v -run "TestTransactionRollback"
```

## Common Issues and Solutions

### Docker Not Running

```bash
# Start Docker Desktop
# On macOS: Open Docker Desktop application
# Verify: docker ps
```

### Container Startup Timeout

```bash
# Increase timeout in test configuration
# Or check Docker resources in Docker Desktop settings
```

### Port Conflicts

```bash
# TestContainers automatically assigns random ports
# No manual configuration needed
```

### Test Data Persistence

```bash
# Tests use cleanup strategies to maintain isolation
# Check defer statements in test functions
```

## Expected Test Results

When running the complete test suite, you should see:

- ✅ All CRUD operations working
- ✅ Error cases properly handled
- ✅ Pattern matching functional
- ✅ Transaction rollback working
- ✅ Test isolation maintained

## Performance Notes

- First test run is slower (Docker image download)
- Subsequent runs are faster (cached images)
- Container startup: ~2-5 seconds
- Test execution: ~1-2 seconds per test

## Learning Outcomes

After completing this practical, you will understand:

- How to set up TestContainers for database testing
- Integration testing best practices
- Test isolation strategies
- Database transaction testing
- CI/CD-ready testing patterns

## Additional Resources

- [TestContainers Go Documentation](https://golang.testcontainers.org/)
- [PostgreSQL TestContainers Module](https://golang.testcontainers.org/modules/postgres/)
- [Go Testing Documentation](https://golang.org/pkg/testing/)

---

## Troubleshooting

If you encounter issues:

1. Ensure Docker Desktop is running
2. Check Go version: `go version`
3. Verify dependencies: `go mod tidy`
4. Check Docker connectivity: `docker ps`
5. Review test output for specific error messages

## Submission

For the practical submission, include:

- Screenshots of test execution
- Code coverage report
- Any challenges faced and how you solved them
