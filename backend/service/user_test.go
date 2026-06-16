package service

import (
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"testing"
)

func initUserTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "s-ui-user-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
}

func TestInitUserStoresHashedDefaultPassword(t *testing.T) {
	initUserTestDB(t)

	user := model.User{}
	if err := database.GetDB().First(&user).Error; err != nil {
		t.Fatalf("load default user: %v", err)
	}
	if user.Password == "admin" {
		t.Fatalf("expected default password to be hashed")
	}
	if !strings.HasPrefix(user.Password, bcryptPasswordPrefix) {
		t.Fatalf("expected bcrypt hash, got %q", user.Password)
	}
}

func TestUserServiceLoginUpgradesLegacyPlaintextPassword(t *testing.T) {
	initUserTestDB(t)

	db := database.GetDB()
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		t.Fatalf("clear users: %v", err)
	}
	legacy := model.User{Username: "legacy", Password: "plain-secret"}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatalf("create legacy user: %v", err)
	}

	service := UserService{}
	username, err := service.Login("legacy", "plain-secret", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if username != "legacy" {
		t.Fatalf("expected legacy username, got %q", username)
	}

	upgraded := model.User{}
	if err := db.First(&upgraded, legacy.Id).Error; err != nil {
		t.Fatalf("load upgraded user: %v", err)
	}
	if upgraded.Password == "plain-secret" {
		t.Fatalf("expected plaintext password to be upgraded")
	}
	if !strings.HasPrefix(upgraded.Password, bcryptPasswordPrefix) {
		t.Fatalf("expected bcrypt hash, got %q", upgraded.Password)
	}
}

func TestUserServiceUpdateAndChangePassStoreHashes(t *testing.T) {
	initUserTestDB(t)

	service := UserService{}
	if err := service.UpdateFirstUser("admin2", "first-secret"); err != nil {
		t.Fatalf("UpdateFirstUser() error = %v", err)
	}
	user, err := service.GetFirstUser()
	if err != nil {
		t.Fatalf("GetFirstUser() error = %v", err)
	}
	if user.Password == "first-secret" || !strings.HasPrefix(user.Password, bcryptPasswordPrefix) {
		t.Fatalf("expected first password to be hashed, got %q", user.Password)
	}

	if err := service.ChangePass("1", "first-secret", "admin3", "second-secret"); err != nil {
		t.Fatalf("ChangePass() error = %v", err)
	}
	changed, err := service.GetFirstUser()
	if err != nil {
		t.Fatalf("GetFirstUser() after change error = %v", err)
	}
	if changed.Username != "admin3" {
		t.Fatalf("expected username admin3, got %q", changed.Username)
	}
	if changed.Password == "second-secret" || !strings.HasPrefix(changed.Password, bcryptPasswordPrefix) {
		t.Fatalf("expected changed password to be hashed, got %q", changed.Password)
	}
	if _, err := service.Login("admin3", "second-secret", "127.0.0.1"); err != nil {
		t.Fatalf("Login() with changed password error = %v", err)
	}
}
