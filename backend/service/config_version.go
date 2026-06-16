package service

import (
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

type ConfigVersionService struct {
}

func (s *ConfigVersionService) GetConfigVersions(nodeId uint, limit int) ([]model.ConfigVersion, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	db := database.GetDB()
	versions := []model.ConfigVersion{}
	query := db.Model(model.ConfigVersion{}).Order("id desc").Limit(limit)
	if nodeId > 0 {
		query = query.Where("node_id = ?", nodeId)
	}
	err := query.Scan(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (s *ConfigVersionService) SaveConfigVersion(version *model.ConfigVersion) error {
	version.Scope = strings.TrimSpace(version.Scope)
	version.Status = strings.TrimSpace(version.Status)
	version.Sha256 = strings.TrimSpace(version.Sha256)

	if version.Scope == "" {
		return fmt.Errorf("config scope is required")
	}
	if version.Status == "" {
		version.Status = "draft"
	}
	if version.CreatedAt == 0 {
		version.CreatedAt = time.Now().Unix()
	}

	db := database.GetDB()
	return db.Save(version).Error
}

func (s *ConfigVersionService) GetConfigDeployments(configVersionId uint64, nodeId uint) ([]model.ConfigDeployment, error) {
	db := database.GetDB()
	deployments := []model.ConfigDeployment{}
	query := db.Model(model.ConfigDeployment{}).Order("id desc")
	if configVersionId > 0 {
		query = query.Where("config_version_id = ?", configVersionId)
	}
	if nodeId > 0 {
		query = query.Where("node_id = ?", nodeId)
	}
	err := query.Scan(&deployments).Error
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

func (s *ConfigVersionService) SaveConfigDeployment(deployment *model.ConfigDeployment) error {
	if deployment.ConfigVersionId == 0 {
		return fmt.Errorf("config version id is required")
	}
	if deployment.NodeId == 0 {
		return fmt.Errorf("node id is required")
	}
	deployment.Status = strings.TrimSpace(deployment.Status)
	if deployment.Status == "" {
		deployment.Status = "pending"
	}
	if deployment.StartedAt == 0 {
		deployment.StartedAt = time.Now().Unix()
	}

	db := database.GetDB()
	return db.Save(deployment).Error
}
