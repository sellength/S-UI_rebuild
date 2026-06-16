package config

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed version
var version string

//go:embed name
var name string

//go:embed config.json
var defaultConfig string

type LogLevel string

const (
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
)

func GetVersion() string {
	return strings.TrimSpace(version)
}

func GetName() string {
	return strings.TrimSpace(name)
}

func GetLogLevel() LogLevel {
	if IsDebug() {
		return Debug
	}
	logLevel := os.Getenv("SUI_LOG_LEVEL")
	if logLevel == "" {
		return Info
	}
	return LogLevel(logLevel)
}

func IsDebug() bool {
	return os.Getenv("SUI_DEBUG") == "true"
}

func GetBinFolderPath() string {
	binFolderPath := os.Getenv("SUI_BIN_FOLDER")
	if binFolderPath == "" {
		binFolderPath = "bin"
	}
	return binFolderPath
}

func GetDBFolderPath() string {
	dbFolderPath := os.Getenv("SUI_DB_FOLDER")
	if dbFolderPath == "" {
		dbFolderPath = "/usr/local/s-ui/db"
	}
	return dbFolderPath
}

func GetDBPath() string {
	return fmt.Sprintf("%s/%s.db", GetDBFolderPath(), GetName())
}

func GetDBType() string {
	dbType := strings.ToLower(strings.TrimSpace(os.Getenv("SUI_DB_TYPE")))
	if dbType == "" {
		return "sqlite"
	}
	return dbType
}

func GetPostgresDSN() string {
	return strings.TrimSpace(os.Getenv("SUI_POSTGRES_DSN"))
}

func GetAgentRegisterToken() string {
	return strings.TrimSpace(os.Getenv("SUI_AGENT_REGISTER_TOKEN"))
}

func GetSecretKey() string {
	return strings.TrimSpace(os.Getenv("SUI_SECRET_KEY"))
}

func GetAgentCertDir() string {
	certDir := strings.TrimSpace(os.Getenv("SUI_AGENT_CERT_DIR"))
	if certDir == "" {
		return "/usr/local/s-ui-agent/certs"
	}
	return certDir
}

func GetACMEShPath() string {
	path := strings.TrimSpace(os.Getenv("SUI_ACME_SH"))
	if path == "" {
		return "acme.sh"
	}
	return path
}

func GetACMEHome() string {
	home := strings.TrimSpace(os.Getenv("SUI_ACME_HOME"))
	if home == "" {
		return "/root/.acme.sh"
	}
	return home
}

func GetCertificateWorkDir() string {
	dir := strings.TrimSpace(os.Getenv("SUI_CERTIFICATE_WORK_DIR"))
	if dir == "" {
		return "/usr/local/s-ui/certificates"
	}
	return dir
}

func GetDefaultConfig() string {
	apiEnv := GetEnvApi()
	if len(apiEnv) > 0 {
		return strings.Replace(defaultConfig, "127.0.0.1:1080", apiEnv, 1)
	}
	return defaultConfig
}

func GetEnvApi() string {
	return os.Getenv("SINGBOX_API")
}
