package api

import (
	"net/http"
	"s-ui/config"
	"s-ui/service"

	"github.com/gin-gonic/gin"
)

type AgentAPIHandler struct {
	service.AgentService
}

func NewAgentAPIHandler(g *gin.RouterGroup) {
	a := &AgentAPIHandler{}
	a.initRouter(g)
}

func (a *AgentAPIHandler) initRouter(g *gin.RouterGroup) {
	g.POST("/register", a.register)
	g.POST("/heartbeat", a.heartbeat)
	g.GET("/config/desired", a.desiredConfig)
	g.POST("/config/report", a.reportConfig)
}

func (a *AgentAPIHandler) register(c *gin.Context) {
	registerToken := config.GetAgentRegisterToken()
	if registerToken == "" {
		c.JSON(http.StatusServiceUnavailable, Msg{Success: false, Msg: "agent registration is disabled until SUI_AGENT_REGISTER_TOKEN is configured"})
		return
	}
	if c.GetHeader("X-Register-Token") != registerToken {
		c.JSON(http.StatusUnauthorized, Msg{Success: false, Msg: "invalid register token"})
		return
	}

	req := service.AgentRegisterRequest{}
	err := c.ShouldBind(&req)
	if err == nil {
		_, err = a.AgentService.Register(&req)
	}
	jsonMsg(c, "register", err)
}

func (a *AgentAPIHandler) heartbeat(c *gin.Context) {
	req := service.AgentHeartbeatRequest{}
	err := c.ShouldBind(&req)
	if req.AgentId == "" {
		req.AgentId = c.GetHeader("X-Agent-Id")
	}
	if req.AgentToken == "" {
		req.AgentToken = c.GetHeader("X-Agent-Token")
	}
	if err == nil {
		err = a.AgentService.Heartbeat(&req)
	}
	jsonMsg(c, "heartbeat", err)
}

func (a *AgentAPIHandler) desiredConfig(c *gin.Context) {
	agentId := c.GetHeader("X-Agent-Id")
	agentToken := c.GetHeader("X-Agent-Token")

	config, err := a.AgentService.GetDesiredConfig(agentId, agentToken)
	jsonObj(c, config, err)
}

func (a *AgentAPIHandler) reportConfig(c *gin.Context) {
	req := service.AgentConfigReportRequest{}
	err := c.ShouldBind(&req)
	if req.AgentId == "" {
		req.AgentId = c.GetHeader("X-Agent-Id")
	}
	if req.AgentToken == "" {
		req.AgentToken = c.GetHeader("X-Agent-Token")
	}
	if err == nil {
		err = a.AgentService.ReportConfig(&req)
	}
	jsonMsg(c, "report", err)
}
