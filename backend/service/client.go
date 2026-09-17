package service

import (
	"encoding/json"
	"s-ui/database"
	"s-ui/database/model"
	"s-ui/logger"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ClientService struct {
}

func (s *ClientService) GetAll() ([]model.Client, error) {
	db := database.GetDB()
	clients := []model.Client{}
	err := db.Model(model.Client{}).Scan(&clients).Error
	if err != nil {
		return nil, err
	}
	return clients, nil
}

func (s *ClientService) Save(tx *gorm.DB, changes []model.Changes) error {
	var err error
	var affectedClientIds []uint
	for _, change := range changes {
		client := model.Client{}
		err = json.Unmarshal(change.Obj, &client)
		if err != nil {
			return err
		}
		switch change.Action {
		case "new":
			err = tx.Create(&client).Error
			if err == nil && client.Id > 0 {
				affectedClientIds = append(affectedClientIds, client.Id)
			}
		case "del":
			err = tx.Where("id = ?", change.Index).Delete(model.Client{}).Error
			if err == nil && change.Index > 0 {
				affectedClientIds = append(affectedClientIds, uint(change.Index))
			}
		default:
			err = tx.Save(client).Error
			if err == nil && client.Id > 0 {
				affectedClientIds = append(affectedClientIds, client.Id)
			}
		}
		if err != nil {
			return err
		}
	}
	if len(affectedClientIds) > 0 {
		_ = CascadeUpdateClients(affectedClientIds, "client-update")
	}
	return err
}

func (s *ClientService) DepleteClients() ([]string, []string, error) {
	var err error
	var clients []model.Client
	var changes []model.Changes
	now := time.Now().Unix()
	db := database.GetDB()
	err = db.Model(model.Client{}).Where("enable = true AND ((volume >0 AND up+down > volume) OR (expiry > 0 AND expiry < ?))", now).Scan(&clients).Error
	if err != nil {
		return nil, nil, err
	}

	dt := time.Now().Unix()
	var users, inbounds []string
	var clientIds []uint
	for _, client := range clients {
		logger.Debug("Client ", client.Name, " is going to be disabled")
		users = append(users, client.Name)
		clientIds = append(clientIds, client.Id)
		var userInbounds []string
		json.Unmarshal(client.Inbounds, &userInbounds)
		inbounds = append(inbounds, userInbounds...)
		changes = append(changes, model.Changes{
			DateTime: dt,
			Actor:    "DepleteJob",
			Key:      "clients",
			Action:   "disable",
			Obj:      json.RawMessage("\"" + client.Name + "\""),
		})
	}

	// Save changes
	if len(changes) > 0 {
		err = db.Model(model.Client{}).Where("enable = true AND ((volume >0 AND up+down > volume) OR (expiry > 0 AND expiry < ?))", now).Update("enable", false).Error
		if err != nil {
			return nil, nil, err
		}
		err = db.Model(model.Changes{}).Create(&changes).Error
		if err != nil {
			return nil, nil, err
		}
		LastUpdate = dt
		_ = CascadeUpdateClients(clientIds, "deplete-job")
	}

	return users, inbounds, nil
}

func CascadeUpdateClients(clientIds []uint, actor string) error {
	if len(clientIds) == 0 {
		return nil
	}
	db := database.GetDB()
	var inboundUsers []model.InboundUser
	if err := db.Model(&model.InboundUser{}).Where("client_id IN ?", clientIds).Scan(&inboundUsers).Error; err != nil {
		return err
	}
	if len(inboundUsers) == 0 {
		return nil
	}

	affectedInbounds := make(map[uint]bool)
	for _, iu := range inboundUsers {
		affectedInbounds[iu.InboundId] = true
	}

	affectedNodes := make(map[uint]bool)
	for inboundId := range affectedInbounds {
		var inbound model.DistributedInbound
		if err := db.Model(&model.DistributedInbound{}).Where("id = ?", inboundId).First(&inbound).Error; err == nil {
			inbound.RenderedConfigJson = nil
			_ = db.Save(&inbound).Error
			_, _ = RenderDistributedInbound(inbound.Id)
			affectedNodes[inbound.NodeId] = true
		}
	}

	if strings.TrimSpace(actor) == "" {
		actor = "client-cascade"
	}

	for nodeId := range affectedNodes {
		if _, err := PublishNodeConfigVersion(nodeId, actor); err != nil {
			logger.Warningf("CascadeUpdateClients: publish node %d failed: %v", nodeId, err)
		} else {
			logger.Infof("CascadeUpdateClients: successfully published new config version for node %d", nodeId)
		}
	}
	return nil
}
