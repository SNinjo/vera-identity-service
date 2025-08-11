package test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
)

func SetupPostgresql() (url string, close func() error, err error) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("mock-db"),
		postgres.WithUsername("mock-user"),
		postgres.WithPassword("mock-pass"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start postgres container | %v", err)
	}

	close = func() error {
		err := pgContainer.Terminate(ctx)
		if err != nil {
			return fmt.Errorf("failed to terminate postgres container | %v", err)
		}
		return nil
	}

	url, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		close()
		return "", nil, fmt.Errorf("failed to get connection string | %v", err)
	}

	return url, close, nil
}

func CleanupTables(db *gorm.DB) error {
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("failed to get tables: %v", err)
	}

	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %v", table, err)
		}
	}

	return nil
}
