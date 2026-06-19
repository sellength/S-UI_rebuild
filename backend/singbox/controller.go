package singbox

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"s-ui/config"
	"strings"
)

var serviceName = "sing-box"

type Controller struct {
}

func (s *Controller) GetBinaryName() string {
	return "sing-box"
}

func (s *Controller) GetBinaryPath() string {
	return config.GetBinFolderPath() + "/" + s.GetBinaryName()
}

func (s *Controller) GetConfigPath() string {
	return config.GetBinFolderPath() + "/config.json"
}

func (s *Controller) IsRunning() bool {
	cmd := exec.Command("pgrep", "sing-box")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If pgrep found the Controller, its output will not be empty
	return strings.TrimSpace(string(output)) != ""
}

func (s *Controller) signalSingbox(signal string) error {
	return os.WriteFile(config.GetBinFolderPath()+"/signal", []byte(signal), fs.ModePerm)
}

func (s *Controller) Restart() error {
	return s.signalSingbox("restart")
}

func (s *Controller) Stop() error {
	if !s.IsRunning() {
		return errors.New("Sing-Box is not running")
	}

	return s.signalSingbox("stop")
}

func (s *Controller) Version() string {
	binPath := s.GetBinaryPath()
	if _, err := os.Stat(binPath); err != nil {
		return "unknown"
	}
	cmd := exec.Command(binPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
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
