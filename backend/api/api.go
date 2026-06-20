package api

import (
	"encoding/json"
	"s-ui/database/model"
	"s-ui/logger"
	"s-ui/service"
	"s-ui/util"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	service.SettingService
	service.UserService
	service.ConfigService
	service.ClientService
	service.TlsService
	service.InDataService
	service.PanelService
	service.StatsService
	service.ServerService
	service.NodeService
	service.CertificateService
	service.ConfigVersionService
	service.DistributedInboundService
	service.SubscriptionService
	service.ConfigTemplateService
}

func NewAPIHandler(g *gin.RouterGroup) {
	a := &APIHandler{}
	a.initRouter(g)
}

func (a *APIHandler) initRouter(g *gin.RouterGroup) {
	g.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "login") && !strings.HasSuffix(path, "logout") {
			checkLogin(c)
		}
	})

	g.POST("/:postAction", a.postHandler)
	g.GET("/sse", a.sseHandler)
	g.GET("/:getAction", a.getHandler)
}

func (a *APIHandler) sseHandler(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	ctx := c.Request.Context()

	// Initial immediate push
	status := a.ServerService.GetStatus("cpu,mem,net,sys,sbd")
	onlines, _ := a.StatsService.GetOnlines()
	c.SSEvent("stats", gin.H{
		"status":     status,
		"onlines":    onlines,
		"lastUpdate": service.LastUpdate,
	})
	c.Writer.Flush()

	for {
		select {
		case <-ctx.Done():
			logger.Info("SSE client connection closed")
			return
		case <-ticker.C:
			status := a.ServerService.GetStatus("cpu,mem,net,sys,sbd")
			onlines, _ := a.StatsService.GetOnlines()

			c.SSEvent("stats", gin.H{
				"status":     status,
				"onlines":    onlines,
				"lastUpdate": service.LastUpdate,
			})
			c.Writer.Flush()
		}
	}
}

func (a *APIHandler) postHandler(c *gin.Context) {
	var err error
	action := c.Param("postAction")
	remoteIP := getRemoteIp(c)

	switch action {
	case "login":
		loginUser, err := a.UserService.Login(c.Request.FormValue("user"), c.Request.FormValue("pass"), remoteIP)
		if err != nil {
			jsonMsg(c, "", err)
			return
		}

		sessionMaxAge, err := a.SettingService.GetSessionMaxAge()
		if err != nil {
			logger.Infof("Unable to get session's max age from DB")
		}

		if sessionMaxAge > 0 {
			err = SetMaxAge(c, sessionMaxAge*60)
			if err != nil {
				logger.Infof("Unable to set session's max age")
			}
		}

		err = SetLoginUser(c, loginUser)
		if err == nil {
			logger.Info("user ", loginUser, " login success")
		} else {
			logger.Warning("login failed: ", err)
		}

		jsonMsg(c, "", nil)
	case "changePass":
		id := c.Request.FormValue("id")
		oldPass := c.Request.FormValue("oldPass")
		newUsername := c.Request.FormValue("newUsername")
		newPass := c.Request.FormValue("newPass")
		err = a.UserService.ChangePass(id, oldPass, newUsername, newPass)
		if err == nil {
			logger.Info("change user credentials success")
			jsonMsg(c, "save", nil)
		} else {
			logger.Warning("change user credentials failed:", err)
			jsonMsg(c, "", err)
		}
	case "save":
		loginUser := GetLoginUser(c)
		data := map[string]string{}
		err = c.ShouldBind(&data)
		if err == nil {
			err = a.ConfigService.SaveChanges(data, loginUser)
		}
		jsonMsg(c, "save", err)
	case "restartApp":
		err = a.PanelService.RestartPanel(3)
		jsonMsg(c, "restartApp", err)
	case "restartSingbox":
		err = a.ServerService.Restart()
		jsonMsg(c, "restartSingbox", err)
	case "stopSingbox":
		err = a.ServerService.Stop()
		jsonMsg(c, "stopSingbox", err)
	case "linkConvert":
		link := c.Request.FormValue("link")
		result, _, err := util.GetOutbound(link, 0)
		jsonObj(c, result, err)
	case "saveNode":
		node := model.Node{}
		err = c.ShouldBind(&node)
		if err == nil {
			err = a.NodeService.Save(&node)
		}
		jsonMsg(c, "save", err)
	case "deleteNode":
		id, convErr := strconv.Atoi(c.Request.FormValue("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		err = a.NodeService.Delete(uint(id))
		jsonMsg(c, "delete", err)
	case "saveDNSProvider":
		provider := model.DNSProvider{
			Id:                   formUint(c, "id"),
			Enable:               formBool(c, "enable", true),
			Name:                 c.Request.FormValue("name"),
			Type:                 c.Request.FormValue("type"),
			CredentialsEncrypted: c.Request.FormValue("credentialsEncrypted"),
			Config:               rawFormJSON(c, "config", `{}`),
		}
		err = a.CertificateService.SaveDNSProvider(&provider)
		jsonMsg(c, "save", err)
	case "deleteDNSProvider":
		id, convErr := strconv.Atoi(c.Request.FormValue("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		err = a.CertificateService.DeleteDNSProvider(uint(id))
		jsonMsg(c, "delete", err)
	case "saveCertificate":
		certificate := model.Certificate{
			Id:              formUint(c, "id"),
			Enable:          formBool(c, "enable", true),
			Name:            c.Request.FormValue("name"),
			Source:          c.Request.FormValue("source"),
			Domains:         rawFormJSON(c, "domains", `[]`),
			Config:          rawFormJSON(c, "config", `{}`),
			Wildcard:        formBool(c, "wildcard", false),
			AutoRenew:       formBool(c, "autoRenew", false),
			DNSProviderId:   formUint(c, "dnsProviderId"),
			ActiveVersionId: formUint(c, "activeVersionId"),
		}
		err = a.CertificateService.SaveCertificate(&certificate)
		jsonMsg(c, "save", err)
	case "deleteCertificate":
		id, convErr := strconv.Atoi(c.Request.FormValue("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		err = a.CertificateService.DeleteCertificate(uint(id))
		jsonMsg(c, "delete", err)
	case "saveCertificateVersion":
		version := model.CertificateVersion{}
		err = c.ShouldBind(&version)
		if err == nil {
			err = a.CertificateService.SaveCertificateVersion(&version)
		}
		jsonMsg(c, "save", err)
	case "issueCertificate":
		id, convErr := strconv.Atoi(c.Request.FormValue("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		version, issueErr := a.CertificateService.IssueCertificate(uint(id))
		jsonObj(c, version, issueErr)
	case "saveConfigVersion":
		version := model.ConfigVersion{}
		err = c.ShouldBind(&version)
		if err == nil {
			err = a.ConfigVersionService.SaveConfigVersion(&version)
		}
		jsonMsg(c, "save", err)
	case "saveConfigDeployment":
		deployment := model.ConfigDeployment{}
		err = c.ShouldBind(&deployment)
		if err == nil {
			err = a.ConfigVersionService.SaveConfigDeployment(&deployment)
		}
		jsonMsg(c, "save", err)
	case "saveNodeConfigTemplates":
		templates := service.NodeConfigTemplates{
			Log:          rawFormJSON(c, "log", `{"level":"info"}`),
			DNS:          rawFormJSON(c, "dns", `{}`),
			Outbounds:    rawFormJSON(c, "outbounds", `[]`),
			Route:        rawFormJSON(c, "route", `{}`),
			Experimental: rawFormJSON(c, "experimental", `{}`),
		}
		err = a.ConfigTemplateService.SaveNodeConfigTemplates(&templates)
		jsonMsg(c, "save", err)
	case "saveDistributedInbound":
		inbound := model.DistributedInbound{
			Id:                    formUint(c, "id"),
			Enable:                formBool(c, "enable", true),
			NodeId:                formUint(c, "nodeId"),
			Protocol:              c.Request.FormValue("protocol"),
			Tag:                   c.Request.FormValue("tag"),
			PublicHost:            c.Request.FormValue("publicHost"),
			Listen:                c.Request.FormValue("listen"),
			ListenPort:            formUint(c, "listenPort"),
			TemplateId:            formUint(c, "templateId"),
			TlsProfileId:          formUint(c, "tlsProfileId"),
			CertificateId:         formUint(c, "certificateId"),
			FormValuesJson:        rawFormJSON(c, "formValuesJson", `{}`),
			AdvancedOverridesJson: rawFormJSON(c, "advancedOverridesJson", `{}`),
			PolicyOverridesJson:   rawFormJSON(c, "policyOverridesJson", `{}`),
		}
		err = a.DistributedInboundService.SaveDistributedInbound(&inbound)
		jsonMsg(c, "save", err)
	case "deleteDistributedInbound":
		id, convErr := strconv.Atoi(c.Request.FormValue("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		err = a.DistributedInboundService.DeleteDistributedInbound(uint(id))
		jsonMsg(c, "delete", err)
	case "saveInboundUser":
		user := model.InboundUser{
			Id:        formUint(c, "id"),
			InboundId: formUint(c, "inboundId"),
			ClientId:  formUint(c, "clientId"),
			Name:      c.Request.FormValue("name"),
			Password:  c.Request.FormValue("password"),
		}
		err = a.DistributedInboundService.SaveInboundUser(&user)
		jsonMsg(c, "save", err)
	case "deleteInboundUser":
		id, convErr := strconv.Atoi(c.Request.FormValue("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		err = a.DistributedInboundService.DeleteInboundUser(uint(id))
		jsonMsg(c, "delete", err)
	case "renderDistributedAnyTLSInbound":
		id, convErr := strconv.Atoi(c.Request.FormValue("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		result, err := service.RenderDistributedInbound(uint(id))
		jsonObj(c, result, err)
	case "publishNodeConfig":
		nodeId, convErr := strconv.Atoi(c.Request.FormValue("nodeId"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		result, err := service.PublishNodeConfigVersion(uint(nodeId), GetLoginUser(c))
		jsonObj(c, result, err)
	case "createSubscription":
		clientId, convErr := strconv.Atoi(c.Request.FormValue("clientId"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		result, err := a.SubscriptionService.CreateSubscription(uint(clientId))
		jsonObj(c, result, err)
	case "deleteSubscription":
		id, convErr := strconv.Atoi(c.Query("id"))
		if convErr != nil {
			jsonMsg(c, "", convErr)
			return
		}
		err := a.SubscriptionService.DeleteSubscription(uint(id))
		jsonMsg(c, "delete subscription", err)
	default:
		jsonMsg(c, "API call", nil)
	}
}

func (a *APIHandler) getHandler(c *gin.Context) {
	action := c.Param("getAction")

	switch action {
	case "logout":
		loginUser := GetLoginUser(c)
		if loginUser != "" {
			logger.Infof("user %s logout", loginUser)
		}
		ClearSession(c)
		jsonMsg(c, "", nil)
	case "load":
		data, err := a.loadData(c)
		if err != nil {
			jsonMsg(c, "", err)
			return
		}
		jsonObj(c, data, nil)
	case "users":
		users, err := a.UserService.GetUsers()
		if err != nil {
			jsonMsg(c, "", err)
			return
		}
		jsonObj(c, *users, nil)
	case "setting":
		data, err := a.SettingService.GetAllSetting()
		if err != nil {
			jsonMsg(c, "", err)
			return
		}
		jsonObj(c, data, err)
	case "stats":
		resource := c.Query("resource")
		tag := c.Query("tag")
		limit, err := strconv.Atoi(c.Query("limit"))
		if err != nil {
			limit = 100
		}
		data, err := a.StatsService.GetStats(resource, tag, limit)
		if err != nil {
			jsonMsg(c, "", err)
			return
		}
		jsonObj(c, data, err)
	case "status":
		request := c.Query("r")
		result := a.ServerService.GetStatus(request)
		jsonObj(c, result, nil)
	case "onlines":
		onlines, err := a.StatsService.GetOnlines()
		jsonObj(c, onlines, err)
	case "logs":
		service := c.Query("s")
		count := c.Query("c")
		level := c.Query("l")
		logs := a.ServerService.GetLogs(service, count, level)
		jsonObj(c, logs, nil)
	case "changes":
		actor := c.Query("a")
		chngKey := c.Query("k")
		count := c.Query("c")
		changes := a.ConfigService.GetChanges(actor, chngKey, count)
		jsonObj(c, changes, nil)
	case "keypairs":
		kType := c.Query("k")
		options := c.Query("o")
		keypair := a.ServerService.GenKeypair(kType, options)
		jsonObj(c, keypair, nil)
	case "nodes":
		nodes, err := a.NodeService.GetAll()
		jsonObj(c, nodes, err)
	case "dnsProviders":
		providers, err := a.CertificateService.GetDNSProviders()
		jsonObj(c, providers, err)
	case "certificates":
		certificates, err := a.CertificateService.GetCertificates()
		jsonObj(c, certificates, err)
	case "certificateVersions":
		id, _ := strconv.Atoi(c.Query("certificateId"))
		versions, err := a.CertificateService.GetCertificateVersions(uint(id))
		jsonObj(c, versions, err)
	case "configVersions":
		nodeId, _ := strconv.Atoi(c.Query("nodeId"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		versions, err := a.ConfigVersionService.GetConfigVersions(uint(nodeId), limit)
		jsonObj(c, versions, err)
	case "configDeployments":
		configVersionId, _ := strconv.ParseUint(c.Query("configVersionId"), 10, 64)
		nodeId, _ := strconv.Atoi(c.Query("nodeId"))
		deployments, err := a.ConfigVersionService.GetConfigDeployments(configVersionId, uint(nodeId))
		jsonObj(c, deployments, err)
	case "nodeConfigTemplates":
		templates, err := a.ConfigTemplateService.GetNodeConfigTemplates()
		jsonObj(c, templates, err)
	case "distributedInbounds":
		nodeId, _ := strconv.Atoi(c.Query("nodeId"))
		inbounds, err := a.DistributedInboundService.GetDistributedInbounds(uint(nodeId))
		jsonObj(c, inbounds, err)
	case "inboundUsers":
		inboundId, _ := strconv.Atoi(c.Query("inboundId"))
		users, err := a.DistributedInboundService.GetInboundUsers(uint(inboundId))
		jsonObj(c, users, err)
	case "subscriptions":
		clientId, _ := strconv.Atoi(c.Query("clientId"))
		subscriptions, err := a.SubscriptionService.GetSubscriptions(uint(clientId))
		jsonObj(c, subscriptions, err)
	case "sse":
		a.sseHandler(c)
	default:
		jsonMsg(c, "API call", nil)
	}
}

func (a *APIHandler) loadData(c *gin.Context) (interface{}, error) {
	data := make(map[string]interface{}, 0)
	lu := c.Query("lu")
	isUpdated, err := a.ConfigService.CheckChanges(lu)
	if err != nil {
		return "", err
	}
	onlines, err := a.StatsService.GetOnlines()



	if err != nil {
		return "", err
	}
	if isUpdated {
		config, err := a.ConfigService.GetConfig()
		if err != nil {
			return "", err
		}
		clients, err := a.ClientService.GetAll()
		if err != nil {
			return "", err
		}
		tlsConfigs, err := a.TlsService.GetAll()
		if err != nil {
			return "", err
		}
		inData, err := a.InDataService.GetAll()
		if err != nil {
			return "", err
		}
		subURI, err := a.SettingService.GetFinalSubURI(strings.Split(c.Request.Host, ":")[0])
		if err != nil {
			return "", err
		}
		data["config"] = *config
		data["clients"] = clients
		data["tls"] = tlsConfigs
		data["inData"] = inData
		data["subURI"] = subURI
		data["onlines"] = onlines
	} else {
		data["onlines"] = onlines
	}

	return data, nil
}

func rawFormJSON(c *gin.Context, key string, fallback string) json.RawMessage {
	value := strings.TrimSpace(c.Request.FormValue(key))
	if value == "" {
		value = fallback
	}
	return json.RawMessage(value)
}

func formUint(c *gin.Context, key string) uint {
	value, _ := strconv.ParseUint(c.Request.FormValue(key), 10, 64)
	return uint(value)
}

func formBool(c *gin.Context, key string, fallback bool) bool {
	value := strings.TrimSpace(c.Request.FormValue(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
