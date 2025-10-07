package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Global test database connection
var testDB *sql.DB

// TestMain sets up the test environment
// This runs ONCE before all tests in this package
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create a PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts("../migrations/init.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start container: %v\n", err)
		os.Exit(1)
	}

	// Ensure container is terminated at the end
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to terminate container: %v\n", err)
		}
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	// Connect to the database
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Verify connection
	if err = testDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

// TestGetByID tests retrieving a user by ID
func TestGetByID(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Test case 1: User exists (from init.sql)
	t.Run("User Exists", func(t *testing.T) {
		user, err := repo.GetByID(1)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if user.Email != "alice@example.com" {
			t.Errorf("Expected email 'alice@example.com', got: %s", user.Email)
		}

		if user.Name != "Alice Smith" {
			t.Errorf("Expected name 'Alice Smith', got: %s", user.Name)
		}
	})

	// Test case 2: User does not exist
	t.Run("User Not Found", func(t *testing.T) {
		_, err := repo.GetByID(9999)
		if err == nil {
			t.Fatal("Expected error for non-existent user, got nil")
		}
	})
}

// TestGetByEmail tests retrieving a user by email
func TestGetByEmail(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("User Exists", func(t *testing.T) {
		user, err := repo.GetByEmail("bob@example.com")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if user.Name != "Bob Johnson" {
			t.Errorf("Expected name 'Bob Johnson', got: %s", user.Name)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		_, err := repo.GetByEmail("nonexistent@example.com")
		if err == nil {
			t.Fatal("Expected error for non-existent email, got nil")
		}
	})
}

// TestCreate tests user creation
func TestCreate(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Create New User", func(t *testing.T) {
		user, err := repo.Create("charlie@example.com", "Charlie Brown")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if user.ID == 0 {
			t.Error("Expected non-zero ID for created user")
		}

		if user.Email != "charlie@example.com" {
			t.Errorf("Expected email 'charlie@example.com', got: %s", user.Email)
		}

		if user.CreatedAt.IsZero() {
			t.Error("Expected non-zero created_at timestamp")
		}

		// Cleanup: delete the created user
		defer repo.Delete(user.ID)
	})

	t.Run("Create Duplicate Email", func(t *testing.T) {
		// Try to create user with existing email (from init.sql)
		_, err := repo.Create("alice@example.com", "Another Alice")
		if err == nil {
			t.Fatal("Expected error when creating user with duplicate email")
		}
	})
}

// TestUpdate tests user updates
func TestUpdate(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Update Existing User", func(t *testing.T) {
		// First, create a user to update
		user, err := repo.Create("david@example.com", "David Davis")
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		defer repo.Delete(user.ID)

		// Update the user
		err = repo.Update(user.ID, "david.updated@example.com", "David Updated")
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// Verify the update
		updatedUser, err := repo.GetByID(user.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated user: %v", err)
		}

		if updatedUser.Email != "david.updated@example.com" {
			t.Errorf("Expected email 'david.updated@example.com', got: %s", updatedUser.Email)
		}

		if updatedUser.Name != "David Updated" {
			t.Errorf("Expected name 'David Updated', got: %s", updatedUser.Name)
		}
	})

	t.Run("Update Non-Existent User", func(t *testing.T) {
		err := repo.Update(9999, "nobody@example.com", "Nobody")
		if err == nil {
			t.Fatal("Expected error when updating non-existent user")
		}
	})
}

// TestDelete tests user deletion
func TestDelete(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Delete Existing User", func(t *testing.T) {
		// Create a user to delete
		user, err := repo.Create("temp@example.com", "Temporary User")
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		// Delete the user
		err = repo.Delete(user.ID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify deletion
		_, err = repo.GetByID(user.ID)
		if err == nil {
			t.Fatal("Expected error when retrieving deleted user")
		}
	})

	t.Run("Delete Non-Existent User", func(t *testing.T) {
		err := repo.Delete(9999)
		if err == nil {
			t.Fatal("Expected error when deleting non-existent user")
		}
	})
}

// TestList tests listing all users
func TestList(t *testing.T) {
	repo := NewUserRepository(testDB)

	users, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	// Should have at least 2 users from init.sql
	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got: %d", len(users))
	}

	// Verify first user
	if users[0].Email != "alice@example.com" {
		t.Errorf("Expected first user email 'alice@example.com', got: %s", users[0].Email)
	}
}

// TestCreateUser_TableDriven demonstrates table-driven testing
func TestCreateUser_TableDriven(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		userName    string
		expectError bool
	}{
		{"Valid User", "valid@example.com", "Valid User", false},
		{"Duplicate Email", "alice@example.com", "Duplicate", true},
		// Note: Empty email test removed as PostgreSQL allows empty strings by default
	}

	repo := NewUserRepository(testDB)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := repo.Create(tc.email, tc.userName)

			if tc.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Cleanup if user was created successfully
			if !tc.expectError && user != nil {
				defer repo.Delete(user.ID)
			}
		})
	}
}

// TestFindByNamePattern tests pattern-based user search
func TestFindByNamePattern(t *testing.T) {
	repo := NewUserRepository(testDB)

	t.Run("Find Users with Smith", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%Smith%")
		if err != nil {
			t.Fatalf("Failed to find users by pattern: %v", err)
		}

		if len(users) == 0 {
			t.Error("Expected to find at least one user with 'Smith' in name")
		}

		// Should find Alice Smith from init.sql
		found := false
		for _, user := range users {
			if user.Email == "alice@example.com" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find Alice Smith in pattern search results")
		}
	})

	t.Run("Find Users with Non-Matching Pattern", func(t *testing.T) {
		users, err := repo.FindByNamePattern("%NonExistent%")
		if err != nil {
			t.Fatalf("Failed to find users by pattern: %v", err)
		}

		if len(users) != 0 {
			t.Errorf("Expected no users for non-matching pattern, got: %d", len(users))
		}
	})
}

// TestCountUsers tests user counting
func TestCountUsers(t *testing.T) {
	repo := NewUserRepository(testDB)

	count, err := repo.CountUsers()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	// Should have at least 2 users from init.sql
	if count < 2 {
		t.Errorf("Expected at least 2 users, got: %d", count)
	}
}

// TestGetRecentUsers tests retrieving recent users
func TestGetRecentUsers(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Create a test user that should be recent
	user, err := repo.Create("recent@example.com", "Recent User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer repo.Delete(user.ID)

	// Get users from last 1 day
	recentUsers, err := repo.GetRecentUsers(1)
	if err != nil {
		t.Fatalf("Failed to get recent users: %v", err)
	}

	// Should include all users since they were just created in init.sql
	if len(recentUsers) == 0 {
		t.Error("Expected to find recent users")
	}

	// Verify our test user is in the results
	found := false
	for _, recentUser := range recentUsers {
		if recentUser.Email == "recent@example.com" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find our test user in recent users")
	}
}

// TestTransactionRollback demonstrates transaction testing
func TestTransactionRollback(t *testing.T) {
	repo := NewUserRepository(testDB)

	// Count users before
	countBefore, err := repo.CountUsers()
	if err != nil {
		t.Fatal(err)
	}

	// Start a transaction that will fail
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// Create user in transaction
	_, err = tx.Exec("INSERT INTO users (email, name) VALUES ($1, $2)",
		"tx@example.com", "TX User")
	if err != nil {
		t.Fatal(err)
	}

	// Rollback transaction
	tx.Rollback()

	// Verify count is unchanged
	countAfter, err := repo.CountUsers()
	if err != nil {
		t.Fatal(err)
	}

	if countAfter != countBefore {
		t.Error("Transaction was not rolled back properly")
	}
}

// TestWithCleanup demonstrates using t.Cleanup
func TestWithCleanup(t *testing.T) {
	repo := NewUserRepository(testDB)

	user, err := repo.Create("cleanup@example.com", "Cleanup User")
	if err != nil {
		t.Fatal(err)
	}

	// t.Cleanup runs after test, even if test fails
	t.Cleanup(func() {
		repo.Delete(user.ID)
	})

	// Test logic - verify user exists
	retrievedUser, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve created user: %v", err)
	}

	if retrievedUser.Email != "cleanup@example.com" {
		t.Errorf("Expected email 'cleanup@example.com', got: %s", retrievedUser.Email)
	}
}