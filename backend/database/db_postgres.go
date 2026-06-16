//go:build postgres

package database

import (
	"fmt"
	"s-ui/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openPostgres(c *gorm.Config) (*gorm.DB, error) {
	dsn := config.GetPostgresDSN()
	if dsn == "" {
		return nil, fmt.Errorf("SUI_POSTGRES_DSN is required when SUI_DB_TYPE=%s", config.GetDBType())
	}
	return gorm.Open(postgres.Open(dsn), c)
}
