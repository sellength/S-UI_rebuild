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

	now := time.Now().Unix()
	for i := range nodes {
		if !nodes[i].Enable {
			continue
		}
		if nodes[i].LastSeenAt <= 0 {
			nodes[i].AgentStatus = "unregistered"
			continue
		}
		diff := now - nodes[i].LastSeenAt
		if diff > 75 {
			nodes[i].AgentStatus = "offline"
			nodes[i].SingboxStatus = "unknown"
		} else if diff > 45 {
			nodes[i].AgentStatus = "timeout"
		}
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
