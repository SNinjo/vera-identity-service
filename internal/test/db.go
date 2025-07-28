package test

import (
	"context"
	"testing"
	"vera-identity-service/internal/db"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupDB(t *testing.T, models ...interface{}) (terminate func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("mock-db"),
		postgres.WithUsername("mock-user"),
		postgres.WithPassword("mock-pass"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	terminate = func() {
		err := pgContainer.Terminate(ctx)
		if err != nil {
			t.Fatalf("failed to terminate postgres container: %v", err)
		}
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		terminate()
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := db.Init(dsn)
	if err != nil {
		terminate()
		t.Fatalf("failed to initialize database: %v", err)
	}

	for _, model := range models {
		err = db.AutoMigrate(model)
		if err != nil {
			terminate()
			t.Fatalf("failed to migrate database: %v", err)
		}
	}

	return terminate
}
