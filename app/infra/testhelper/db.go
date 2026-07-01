// infra/testhelper/db.go
package testhelper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// CI — use existing postgres service
	if dsn := os.Getenv("TEST_DB_DSN"); dsn != "" {
		db, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			t.Fatalf("failed to connect to CI db: %v", err)
		}
		runMigrations(t, db)
		t.Cleanup(func() { cleanDB(t, db) })
		return db
	}

	return newContainerDB(t)
}

func newContainerDB(t *testing.T) *sqlx.DB {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "testdb",
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	runMigrations(t, db)

	t.Cleanup(func() {
		db.Close()
		container.Terminate(ctx)
	})

	return db
}

func runMigrations(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// get absolute path to this file → resolve migrations relative to it
	_, filename, _, _ := runtime.Caller(0)
	// testhelper is at infra/testhelper/db.go
	// migrations are at migrations/versioned/
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations", "versioned")
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no migration files found at %s", migrationsDir)
	}

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", f, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			t.Fatalf("failed to run migration %s: %v", f, err)
		}
	}
}

func cleanDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
	// truncate all tables in reverse dependency order
	_, err := db.Exec(`
        TRUNCATE TABLE 
            sensor_gateway_mapping,
            sensors,
            downlink_requests
        RESTART IDENTITY CASCADE
    `)
	if err != nil {
		t.Logf("failed to clean db: %v", err)
	}
}
