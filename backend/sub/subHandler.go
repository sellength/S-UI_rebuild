package sub

import (
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"s-ui/logger"
	"s-ui/service"

	"github.com/gin-gonic/gin"
)

type SubHandler struct {
	service.SettingService
	SubService
	JsonService
	ClashService
}

func NewSubHandler(g *gin.RouterGroup) {
	a := &SubHandler{}
	a.initRouter(g)
}

func (s *SubHandler) initRouter(g *gin.RouterGroup) {
	g.GET("/:subid", s.subs)
	g.HEAD("/:subid", s.subHeaders)
}

func (s *SubHandler) subs(c *gin.Context) {
	var headers []string
	var result *string
	var err error
	subId := c.Param("subid")
	format, isFormat := c.GetQuery("format")
	if isFormat {
		switch format {
		case "json":
			result, err = s.JsonService.GetJson(subId, format)
		case "clash":
			result, headers, err = s.ClashService.GetClash(subId)
		}
		if err != nil || result == nil {
			logger.Error(err)
			c.String(400, "Error!")
			return
		}
	} else {
		result, headers, err = s.SubService.GetSubs(subId)
		if err != nil || result == nil {
			logger.Error(err)
			c.String(400, "Error!")
			return
		}
	}

	if len(headers) > 0 {
		s.addHeaders(c, headers)
	}

	c.String(200, *result)
}

func (s *SubHandler) subHeaders(c *gin.Context) {
	subId := c.Param("subid")
	client := &model.Client{}
	db := database.GetDB()
	err := db.Model(model.Client{}).Where("enable = true and name = ?", subId).First(client).Error
	if err != nil {
		logger.Error(err)
		c.String(400, "Error!")
		return
	}

	updateInterval, _ := s.SettingService.GetSubUpdates()
	headers := []string{
		fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", client.Up, client.Down, client.Volume, client.Expiry),
		fmt.Sprintf("%d", updateInterval),
		subId,
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
