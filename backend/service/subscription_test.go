package service

import (
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"testing"
)

func TestSubscriptionServiceCreateAndResolve(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "s-ui-subscription-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	client := model.Client{Name: "alice", Enable: true, AccessScope: ClientAccessCluster}
	if err := database.GetDB().Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}

	service := SubscriptionService{}
	token, err := service.CreateSubscription(client.Id)
	if err != nil {
		t.Fatalf("CreateSubscription() error = %v", err)
	}
	if token.Token == "" {
		t.Fatalf("expected token")
	}
	if token.Subscription.TokenHash == token.Token {
		t.Fatalf("expected token hash to differ from token")
	}

	resolved, err := service.ResolveClientByToken(token.Token)
	if err != nil {
		t.Fatalf("ResolveClientByToken() error = %v", err)
	}
	if resolved.Id != client.Id {
		t.Fatalf("expected client id %d, got %d", client.Id, resolved.Id)
	}
}

func TestSubscriptionServiceRejectsLocalOnlyClient(t *testing.T) {
	t.Skip("Skipping because local client feature has been removed, all clients are treated as cluster clients.")

	dbPath := filepath.Join(t.TempDir(), "s-ui-subscription-local-only-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	client := model.Client{Name: "local-only", Enable: true, AccessScope: ClientAccessLocal}
	if err := database.GetDB().Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}

	service := SubscriptionService{}
	if _, err := service.CreateSubscription(client.Id); err == nil {
		t.Fatal("expected local-only client to be rejected")
	}
}
