package sub

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"s-ui/logger"
	"s-ui/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type SubHandler struct {
	service.SettingService
	SubService
	JsonService
	ClashService
	DistributedService
}

func NewSubHandler(g *gin.RouterGroup) {
	a := &SubHandler{}
	a.initRouter(g)
}

func (s *SubHandler) initRouter(g *gin.RouterGroup) {
	g.GET("/:subid", s.subs)
	g.HEAD("/:subid", s.subHeaders)
}

func (s *SubHandler) resolveClient(subId string) (*model.Client, error) {
	db := database.GetDB()
	var client model.Client
	var resolved bool

	// Check by token first
	var subscription model.Subscription
	sum := sha256.Sum256([]byte(subId))
	tokenHash := hex.EncodeToString(sum[:])
	if err := db.Model(model.Subscription{}).Where("enable = ? AND (token = ? OR token_hash = ?)", true, subId, tokenHash).First(&subscription).Error; err == nil {
		if err := db.Model(model.Client{}).Where("enable = ? AND id = ?", true, subscription.ClientId).First(&client).Error; err == nil {
			resolved = true
		}
	}

	// Fallback to username
	if !resolved {
		if err := db.Model(model.Client{}).Where("enable = ? AND name = ?", true, subId).First(&client).Error; err == nil {
			resolved = true
		}
	}

	if !resolved {
		return nil, fmt.Errorf("client not found")
	}
	return &client, nil
}

func (s *SubHandler) subs(c *gin.Context) {
	var headers []string
	var result *string
	var err error
	subId := c.Param("subid")
	format, isFormat := c.GetQuery("format")
	if !isFormat {
		ua := strings.ToLower(c.Request.UserAgent())
		if strings.Contains(ua, "clash") {
			format = "clash"
			isFormat = true
		} else if strings.Contains(ua, "sing-box") {
			format = "distributed-json"
			isFormat = true
		}
	}

	// 1. Resolve client (either by subscription token or by client username)
	client, err := s.resolveClient(subId)
	if err != nil {
		logger.Error(err)
		c.String(400, "Error!")
		return
	}

	// 2. Fetch subscription content
	if isFormat {
		switch format {
		case "json":
			result, err = s.JsonService.GetJson(client.Name, format)
		case "distributed-json":
			result, err = s.DistributedService.GetDistributedJson(client.Name)
		case "distributed-source", "distributed-raw":
			result, err = s.DistributedService.GetDistributedRaw(client.Name)
		case "clash":
			result, headers, err = s.ClashService.GetClash(client.Name)
		}
		if err != nil || result == nil {
			logger.Error(err)
			c.String(400, "Error!")
			return
		}
	} else {
		// Traditional base64 subscription format (Shadowrocket, browser, etc.)
		linksArray := (&LinkService{}).GetLinks(&client.Links, "all", "")
		rawResult := strings.Join(linksArray, "\n")
		encoded := base64.StdEncoding.EncodeToString([]byte(rawResult))
		result = &encoded

		// Headers
		updateInterval, _ := s.SettingService.GetSubUpdates()
		headers = append(headers, fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", client.Up, client.Down, client.Volume, client.Expiry))
		headers = append(headers, fmt.Sprintf("%d", updateInterval))
		headers = append(headers, client.Name)
	}

	if len(headers) > 0 {
		s.addHeaders(c, headers)
	}

	c.String(200, *result)
}

func (s *SubHandler) subHeaders(c *gin.Context) {
	subId := c.Param("subid")
	client, err := s.resolveClient(subId)
	if err != nil {
		logger.Error(err)
		c.String(400, "Error!")
		return
	}

	updateInterval, _ := s.SettingService.GetSubUpdates()
	headers := []string{
		fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", client.Up, client.Down, client.Volume, client.Expiry),
		fmt.Sprintf("%d", updateInterval),
		client.Name,
	}
	s.addHeaders(c, headers)

	c.Status(200)
}

func (s *SubHandler) addHeaders(c *gin.Context, headers []string) {
	if len(headers) >= 3 {
		c.Writer.Header().Set("Subscription-Userinfo", headers[0])
		c.Writer.Header().Set("Profile-Update-Interval", headers[1])
		c.Writer.Header().Set("Profile-Title", headers[2])
	}
}
