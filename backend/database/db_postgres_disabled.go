//go:build !postgres

package database

import (
	"fmt"

	"gorm.io/gorm"
)

func openPostgres(_ *gorm.Config) (*gorm.DB, error) {
	return nil, fmt.Errorf("postgres support is not built in; rebuild with -tags postgres")
}
