package database

import (
	"fmt"
	"os"
	"path"
	"s-ui/config"
	"s-ui/database/model"

	sqlite "github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func initUser() error {
	var count int64
	err := db.Model(&model.User{}).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		password, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user := &model.User{
			Username: "admin",
			Password: string(password),
		}
		return db.Create(user).Error
	}
	return nil
}

func OpenDB(dbPath string) error {
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, 01740)
	if err != nil {
		return err
	}

	var gormLogger logger.Interface

	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}

	switch config.GetDBType() {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dbPath), c)
	case "postgres", "postgresql":
		db, err = openPostgres(c)
	default:
		return fmt.Errorf("unsupported SUI_DB_TYPE %q", config.GetDBType())
	}
	return err
}

func InitDB(dbPath string) error {
	err := OpenDB(dbPath)
	if err != nil {
		return err
	}

	err = db.AutoMigrate(
		&model.Setting{},
		&model.Tls{},
		&model.InboundData{},
		&model.User{},
		&model.Stats{},
		&model.Client{},
		&model.Changes{},
		&model.Node{},
		&model.NodeAgent{},
		&model.NodeGroup{},
		&model.NodeGroupMember{},
		&model.DNSProvider{},
		&model.Certificate{},
		&model.CertificateVersion{},
		&model.ProtocolTemplate{},
		&model.DistributedInbound{},
		&model.InboundUser{},
		&model.ConfigVersion{},
		&model.ConfigDeployment{},
		&model.NodeHeartbeat{},
		&model.NodeMetric{},
		&model.Subscription{},
		&model.SubStoreIntegration{},
		&model.AuditLog{},
	)
	if err != nil {
		return err
	}
	err = initUser()
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
