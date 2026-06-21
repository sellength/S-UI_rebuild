package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"s-ui/database/model"
	"s-ui/singbox"
)

type agentConfig struct {
	BaseURL       string
	NodeCode      string
	AgentID       string
	AgentToken    string
	RegisterToken string
	Interval      time.Duration
	ConfigDir     string
	CertDir       string
	SingboxBin    string
	CheckConfig   bool
	ReloadCommand string
	StatePath     string
}

type agentState struct {
	LastReportedConfigID uint64 `json:"lastReportedConfigId"`
	LastAppliedVersion   uint64 `json:"lastAppliedVersion"`
	AppliedSha256        string `json:"appliedSha256"`
}

type apiMessage struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj"`
}

type desiredConfig struct {
	NodeID        uint                `json:"nodeId"`
	ConfigVersion *configVersion      `json:"configVersion"`
	Certificates  []certificateBundle `json:"certificates"`
}

type certificateBundle struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Fingerprint   string `json:"fingerprint"`
	FullchainPEM  string `json:"fullchainPem"`
	PrivateKeyPEM string `json:"privateKeyPem"`
	NotBefore     int64  `json:"notBefore"`
	NotAfter      int64  `json:"notAfter"`
}

type configVersion struct {
	ID          uint64          `json:"id"`
	Version     uint64          `json:"version"`
	Status      string          `json:"status"`
	Sha256      string          `json:"sha256"`
	ContentJSON json.RawMessage `json:"contentJson"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fatal(err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	if err := register(client, cfg); err != nil {
		fatal(err)
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	state := loadState(cfg.StatePath)
	for {
		if err := heartbeat(client, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "heartbeat failed: %v\n", err)
		}

		desired, err := getDesiredConfig(client, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "desired config failed: %v\n", err)
		} else if desired.ConfigVersion != nil && desired.ConfigVersion.ID != state.LastReportedConfigID {
			if err := reportConfig(client, cfg, desired.ConfigVersion.ID, "pulled", ""); err != nil {
				fmt.Fprintf(os.Stderr, "config report failed: %v\n", err)
			} else {
				status, errMsg := stageDesiredBundle(cfg, desired)
				if err := reportConfig(client, cfg, desired.ConfigVersion.ID, status, errMsg); err != nil {
					fmt.Fprintf(os.Stderr, "config report failed: %v\n", err)
				}
				if status != "failed" {
					state.LastReportedConfigID = desired.ConfigVersion.ID
					if status == "applied" {
						state.LastAppliedVersion = desired.ConfigVersion.Version
						state.AppliedSha256 = desired.ConfigVersion.Sha256
					}
					if err := saveState(cfg.StatePath, state); err != nil {
						fmt.Fprintf(os.Stderr, "state save failed: %v\n", err)
					}
				}
			}
		}

		<-ticker.C
	}
}

func loadConfig() (*agentConfig, error) {
	defaultAgentID, _ := os.Hostname()
	if defaultAgentID == "" {
		defaultAgentID = "s-ui-agent"
	}

	baseURL := flag.String("base-url", env("SUI_AGENT_BASE_URL", ""), "Control plane agent API base URL, e.g. https://panel.example.com/app/agent")
	nodeCode := flag.String("node-code", env("SUI_NODE_CODE", ""), "Node code created in the panel")
	agentID := flag.String("agent-id", env("SUI_AGENT_ID", defaultAgentID), "Stable agent id")
	agentToken := flag.String("agent-token", env("SUI_AGENT_TOKEN", ""), "Agent shared token")
	registerToken := flag.String("register-token", env("SUI_AGENT_REGISTER_TOKEN", ""), "Registration token required by the control plane")
	interval := flag.Duration("interval", envDuration("SUI_AGENT_INTERVAL", 30*time.Second), "Heartbeat interval")
	configDir := flag.String("config-dir", env("SUI_AGENT_CONFIG_DIR", "/usr/local/s-ui-agent/configs"), "Directory for current, pending, and backup sing-box configs")
	certDir := flag.String("cert-dir", env("SUI_AGENT_CERT_DIR", "/usr/local/s-ui-agent/certs"), "Directory for certificate bundles")
	singboxBin := flag.String("singbox-bin", env("SUI_SINGBOX_BIN", "sing-box"), "sing-box binary path")
	checkConfig := flag.Bool("check-config", envBool("SUI_AGENT_CHECK_CONFIG", true), "Run sing-box check before promoting pending config")
	reloadCommand := flag.String("reload-command", env("SUI_AGENT_RELOAD_COMMAND", ""), "Optional command to apply current config after validation")
	statePath := flag.String("state-path", env("SUI_AGENT_STATE_PATH", "/usr/local/s-ui-agent/state.json"), "Agent state file path")
	flag.Parse()

	cfg := &agentConfig{
		BaseURL:       strings.TrimRight(strings.TrimSpace(*baseURL), "/"),
		NodeCode:      strings.TrimSpace(*nodeCode),
		AgentID:       strings.TrimSpace(*agentID),
		AgentToken:    strings.TrimSpace(*agentToken),
		RegisterToken: strings.TrimSpace(*registerToken),
		Interval:      *interval,
		ConfigDir:     strings.TrimSpace(*configDir),
		CertDir:       strings.TrimSpace(*certDir),
		SingboxBin:    strings.TrimSpace(*singboxBin),
		CheckConfig:   *checkConfig,
		ReloadCommand: strings.TrimSpace(*reloadCommand),
		StatePath:     strings.TrimSpace(*statePath),
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if cfg.NodeCode == "" {
		return nil, fmt.Errorf("node code is required")
	}
	if cfg.AgentID == "" {
		return nil, fmt.Errorf("agent id is required")
	}
	if cfg.AgentToken == "" {
		return nil, fmt.Errorf("agent token is required")
	}
	if cfg.RegisterToken == "" {
		return nil, fmt.Errorf("register token is required")
	}
	if cfg.Interval < 3*time.Second {
		cfg.Interval = 3 * time.Second
	}
	if cfg.ConfigDir == "" {
		return nil, fmt.Errorf("config dir is required")
	}
	if cfg.CertDir == "" {
		return nil, fmt.Errorf("cert dir is required")
	}
	if cfg.StatePath == "" {
		return nil, fmt.Errorf("state path is required")
	}

	return cfg, nil
}

func register(client *http.Client, cfg *agentConfig) error {
	_, singboxVersion := collectSingboxHealth(cfg)
	body := map[string]interface{}{
		"nodeCode":       cfg.NodeCode,
		"agentId":        cfg.AgentID,
		"agentToken":     cfg.AgentToken,
		"agentVersion":   "0.1.0",
		"singboxVersion": singboxVersion,
	}

	headers := map[string]string{}
	if cfg.RegisterToken != "" {
		headers["X-Register-Token"] = cfg.RegisterToken
	}

	_, err := postJSON(client, cfg.BaseURL+"/register", body, headers)
	return err
}

func heartbeat(client *http.Client, cfg *agentConfig) error {
	state := loadState(cfg.StatePath)
	singboxStatus, singboxVersion := collectSingboxHealth(cfg)

	var stats []*model.Stats
	if singboxStatus == "running" {
		var api singbox.V2rayAPI
		apiAddr := os.Getenv("SUI_SINGBOX_API")
		if apiAddr == "" {
			apiAddr = "127.0.0.1:10080"
		}
		if err := api.Init(apiAddr); err == nil {
			if s, err := api.GetStats(true); err == nil {
				stats = s
			}
			api.Close()
		}
	}

	body := map[string]interface{}{
		"configVersion":  state.LastAppliedVersion,
		"agentStatus":    "online",
		"singboxStatus":  singboxStatus,
		"agentVersion":   "0.1.0",
		"singboxVersion": singboxVersion,
		"appliedSha256":  state.AppliedSha256,
		"stats":          stats,
		"resourceSummary": fmt.Sprintf(
			`{"goos":%q,"goarch":%q,"goroutines":%d,"singboxBin":%q}`,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.NumGoroutine(),
			cfg.SingboxBin,
		),
	}

	_, err := postJSON(client, cfg.BaseURL+"/heartbeat", body, agentHeaders(cfg))
	return err
}

func getDesiredConfig(client *http.Client, cfg *agentConfig) (*desiredConfig, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.BaseURL+"/config/desired", nil)
	if err != nil {
		return nil, err
	}
	for k, v := range agentHeaders(cfg) {
		req.Header.Set(k, v)
	}

	msg, err := doRequest(client, req)
	if err != nil {
		return nil, err
	}

	desired := desiredConfig{}
	if len(msg.Obj) == 0 || string(msg.Obj) == "null" {
		return &desired, nil
	}
	if err := json.Unmarshal(msg.Obj, &desired); err != nil {
		return nil, err
	}
	return &desired, nil
}

func reportConfig(client *http.Client, cfg *agentConfig, configVersionID uint64, status string, errMsg string) error {
	body := map[string]interface{}{
		"configVersionId": configVersionID,
		"status":          status,
		"errorMessage":    errMsg,
	}
	_, err := postJSON(client, cfg.BaseURL+"/config/report", body, agentHeaders(cfg))
	return err
}

func stageDesiredBundle(cfg *agentConfig, desired *desiredConfig) (string, string) {
	if desired == nil {
		return "pulled", ""
	}
	if err := writeCertificates(cfg.CertDir, desired.Certificates); err != nil {
		return "failed", err.Error()
	}
	return stageDesiredConfig(cfg, desired.ConfigVersion)
}

func writeCertificates(certDir string, certificates []certificateBundle) error {
	if len(certificates) == 0 {
		return nil
	}
	if err := os.MkdirAll(certDir, 0750); err != nil {
		return err
	}

	for _, certificate := range certificates {
		if certificate.ID == 0 {
			return fmt.Errorf("certificate id is required")
		}
		if strings.TrimSpace(certificate.FullchainPEM) == "" || strings.TrimSpace(certificate.PrivateKeyPEM) == "" {
			return fmt.Errorf("certificate %d pem data is required", certificate.ID)
		}

		name := safeFileName(certificate.Name)
		if name == "" {
			name = fmt.Sprintf("cert-%d", certificate.ID)
		}
		if certificate.Fingerprint != "" {
			name = certificateDirName(certificate.ID, certificate.Name, certificate.Fingerprint)
		}

		targetDir := filepath.Join(certDir, name)
		if err := os.MkdirAll(targetDir, 0700); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(targetDir, "fullchain.pem"), []byte(ensureTrailingNewline(certificate.FullchainPEM)), 0600); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(targetDir, "privkey.pem"), []byte(ensureTrailingNewline(certificate.PrivateKeyPEM)), 0600); err != nil {
			return err
		}
	}

	return nil
}

func certificateDirName(id uint, name string, fingerprint string) string {
	safeName := safeFileName(name)
	if safeName == "" {
		safeName = fmt.Sprintf("cert-%d", id)
	}
	if strings.TrimSpace(fingerprint) == "" {
		return safeName
	}
	return fmt.Sprintf("%s-%s", safeName, shortFingerprint(fingerprint))
}

func stageDesiredConfig(cfg *agentConfig, version *configVersion) (string, string) {
	if version == nil || len(version.ContentJSON) == 0 || string(version.ContentJSON) == "null" {
		return "pulled", ""
	}

	if err := os.MkdirAll(cfg.ConfigDir, 0750); err != nil {
		return "failed", err.Error()
	}

	currentPath := filepath.Join(cfg.ConfigDir, "current.json")
	pendingPath := filepath.Join(cfg.ConfigDir, "pending.json")
	backupPath := filepath.Join(cfg.ConfigDir, "backup.json")

	if err := os.WriteFile(pendingPath, prettyJSON(version.ContentJSON), 0600); err != nil {
		return "failed", err.Error()
	}

	if cfg.CheckConfig {
		if err := checkSingboxConfig(cfg.SingboxBin, pendingPath); err != nil {
			return "failed", err.Error()
		}
	}

	if _, err := os.Stat(currentPath); err == nil {
		if data, err := os.ReadFile(currentPath); err == nil {
			if err := os.WriteFile(backupPath, data, 0600); err != nil {
				return "failed", err.Error()
			}
		}
	}

	if err := os.Rename(pendingPath, currentPath); err != nil {
		return "failed", err.Error()
	}

	if cfg.ReloadCommand == "" {
		return "validated", ""
	}
	if err := runReloadCommand(cfg.ReloadCommand); err != nil {
		return "failed", err.Error()
	}

	return "applied", ""
}

func safeFileName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	builder := strings.Builder{}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			continue
		}
		if r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('-')
	}
	return strings.Trim(builder.String(), "-")
}

func shortFingerprint(value string) string {
	value = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), ":", ""))
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func ensureTrailingNewline(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return value + "\n"
}

func checkSingboxConfig(singboxBin string, configPath string) error {
	cmd := exec.Command(singboxBin, "check", "-c", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sing-box check failed: %v: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runReloadCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reload command failed: %v: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func collectSingboxHealth(cfg *agentConfig) (string, string) {
	return singboxRuntimeStatus(cfg.SingboxBin), singboxVersion(cfg.SingboxBin)
}

func dockerContainerRunning(containerName string) (string, bool) {
	if _, err := exec.LookPath("docker"); err != nil {
		return "", false
	}
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", containerName)
	var out bytes.Buffer
	cmd.Stdout = &out
	if cmd.Run() == nil {
		status := strings.TrimSpace(out.String())
		if status == "running" {
			return "running", true
		}
		if status != "" {
			return status, true
		}
	}
	return "", false
}

func dockerContainerVersion(containerName string) (string, bool) {
	if _, err := exec.LookPath("docker"); err != nil {
		return "", false
	}
	cmd := exec.Command("docker", "exec", containerName, "sing-box", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if cmd.Run() == nil {
		return parseSingboxVersion(out.String()), true
	}
	return "", false
}

func singboxVersion(singboxBin string) string {
	if version, ok := dockerContainerVersion("sing-box"); ok {
		return version
	}
	if !binaryAvailable(singboxBin) {
		return "unknown"
	}
	cmd := exec.Command(singboxBin, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	return parseSingboxVersion(string(output))
}

func parseSingboxVersion(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		for i, field := range fields {
			if strings.EqualFold(field, "version") && i+1 < len(fields) {
				return strings.TrimPrefix(strings.TrimSpace(fields[i+1]), "v")
			}
		}
		if len(fields) == 1 && strings.HasPrefix(fields[0], "v") {
			return strings.TrimPrefix(fields[0], "v")
		}
	}
	return "unknown"
}

func singboxRuntimeStatus(singboxBin string) string {
	if status, ok := dockerContainerRunning("sing-box"); ok {
		return status
	}
	if !binaryAvailable(singboxBin) {
		return "not_installed"
	}
	if status, ok := systemctlSingboxStatus(); ok {
		return status
	}
	if processRunning("sing-box") || processRunning(filepath.Base(singboxBin)) {
		return "running"
	}
	return "stopped"
}

func systemctlSingboxStatus() (string, bool) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return "", false
	}
	
	// 优先检测 S-UI Node Agent 单独的 s-ui-agent-singbox 服务
	cmd := exec.Command("systemctl", "is-active", "s-ui-agent-singbox")
	output, err := cmd.CombinedOutput()
	status := strings.TrimSpace(string(output))
	if err == nil && status == "active" {
		return "running", true
	}

	// 备用检测系统原生的 sing-box 服务
	cmd = exec.Command("systemctl", "is-active", "sing-box")
	output, err = cmd.CombinedOutput()
	status = strings.TrimSpace(string(output))
	if err == nil && status == "active" {
		return "running", true
	}

	switch status {
	case "inactive", "failed", "deactivating", "activating":
		return status, true
	default:
		return "", false
	}
}

func processRunning(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	if _, err := exec.LookPath("pgrep"); err != nil {
		return false
	}
	cmd := exec.Command("pgrep", "-x", name)
	return cmd.Run() == nil
}

func binaryAvailable(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if strings.Contains(path, "/") {
		info, err := os.Stat(path)
		return err == nil && !info.IsDir() && info.Mode()&0111 != 0
	}
	_, err := exec.LookPath(path)
	return err == nil
}

func prettyJSON(raw json.RawMessage) []byte {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return raw
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return raw
	}
	return append(data, '\n')
}

func loadState(path string) agentState {
	data, err := os.ReadFile(path)
	if err != nil {
		return agentState{}
	}
	state := agentState{}
	if err := json.Unmarshal(data, &state); err != nil {
		return agentState{}
	}
	return state
}

func saveState(path string, state agentState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0600)
}

func postJSON(client *http.Client, url string, body interface{}, headers map[string]string) (*apiMessage, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return doRequest(client, req)
}

func doRequest(client *http.Client, req *http.Request) (*apiMessage, error) {
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("%s returned %s: %s", req.URL.String(), res.Status, strings.TrimSpace(string(data)))
	}

	msg := apiMessage{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	if !msg.Success {
		return nil, fmt.Errorf("%s", msg.Msg)
	}
	return &msg, nil
}

func agentHeaders(cfg *agentConfig) map[string]string {
	return map[string]string{
		"X-Agent-Id":    cfg.AgentID,
		"X-Agent-Token": cfg.AgentToken,
	}
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
