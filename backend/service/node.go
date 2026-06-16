package service

import (
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

type NodeService struct {
}

func (s *NodeService) GetAll() ([]model.Node, error) {
	db := database.GetDB()
	nodes := []model.Node{}
	err := db.Model(model.Node{}).Order("id asc").Scan(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (s *NodeService) Save(node *model.Node) error {
	node.Name = strings.TrimSpace(node.Name)
	node.Code = strings.TrimSpace(node.Code)
	node.PublicHost = strings.TrimSpace(node.PublicHost)

	if node.Name == "" {
		return fmt.Errorf("node name is required")
	}
	if node.Code == "" {
		return fmt.Errorf("node code is required")
	}

	now := time.Now().Unix()
	db := database.GetDB()
	if node.Id == 0 {
		node.CreatedAt = now
	} else if node.CreatedAt == 0 {
		existing := model.Node{}
		if err := db.Model(model.Node{}).Where("id = ?", node.Id).First(&existing).Error; err != nil {
			return err
		}
		node.CreatedAt = existing.CreatedAt
	}
	node.UpdatedAt = now

	return db.Save(node).Error
}

func (s *NodeService) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("node id is required")
	}
	db := database.GetDB()
	return db.Where("id = ?", id).Delete(model.Node{}).Error
}
