package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"time"
)

type SubscriptionService struct {
}

type SubscriptionToken struct {
	Token        string             `json:"token"`
	Subscription model.Subscription `json:"subscription"`
}

func (s *SubscriptionService) GetSubscriptions(clientId uint) ([]model.Subscription, error) {
	db := database.GetDB()
	subscriptions := []model.Subscription{}
	query := db.Model(model.Subscription{}).Order("id asc")
	if clientId > 0 {
		query = query.Where("client_id = ?", clientId)
	}
	err := query.Scan(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *SubscriptionService) CreateSubscription(clientId uint) (*SubscriptionToken, error) {
	if clientId == 0 {
		return nil, fmt.Errorf("client id is required")
	}

	db := database.GetDB()
	client := model.Client{}
	if err := db.Model(model.Client{}).Where("id = ?", clientId).First(&client).Error; err != nil {
		return nil, err
	}
	if !clientAllowsCluster(client) {
		return nil, fmt.Errorf("client is not allowed to use cluster subscriptions")
	}

	token, err := randomToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	subscription := model.Subscription{
		Enable:    true,
		ClientId:  clientId,
		TokenHash: hashSubscriptionToken(token),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&subscription).Error; err != nil {
		return nil, err
	}

	return &SubscriptionToken{
		Token:        token,
		Subscription: subscription,
	}, nil
}

func (s *SubscriptionService) ResolveClientByToken(token string) (*model.Client, error) {
	hash := hashSubscriptionToken(token)
	db := database.GetDB()
	subscription := model.Subscription{}
	if err := db.Model(model.Subscription{}).Where("enable = ? AND token_hash = ?", true, hash).First(&subscription).Error; err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	subscription.LastUsedAt = now
	_ = db.Save(&subscription).Error

	client := &model.Client{}
	if err := db.Model(model.Client{}).Where("enable = ? AND id = ?", true, subscription.ClientId).First(client).Error; err != nil {
		return nil, err
	}
	return client, nil
}

func randomToken() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func hashSubscriptionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *SubscriptionService) DeleteSubscription(id uint) error {
	if id == 0 {
		return fmt.Errorf("subscription id is required")
	}
	db := database.GetDB()
	return db.Where("id = ?", id).Delete(&model.Subscription{}).Error
}
