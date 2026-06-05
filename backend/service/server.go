package service

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"s-ui/config"
	"s-ui/logger"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	publicIPv4 = "Fetching..."
	publicIPv6 = "Fetching..."
)

func init() {
	go refreshPublicIPsLoop()
}

func refreshPublicIPsLoop() {
	time.Sleep(2 * time.Second) // 避开启动网络波动
	refreshPublicIPs()

	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		refreshPublicIPs()
	}
}

func refreshPublicIPs() {
	go func() {
		providers := []string{
			"http://api.ipify.org",
			"http://ip4.ident.me",
			"http://icanhazip.com",
		}
		client := &http.Client{Timeout: 5 * time.Second}
		for _, provider := range providers {
			resp, err := client.Get(provider)
			if err == nil {
				body, readErr := io.ReadAll(resp.Body)
				resp.Body.Close()
				if readErr == nil {
					ip := strings.TrimSpace(string(body))
					if len(ip) > 0 {
						publicIPv4 = ip
						return
					}
				}
			}
		}
		publicIPv4 = "N/A"
	}()

	go func() {
		providers := []string{
			"http://api6.ipify.org",
			"http://ip6.ident.me",
			"http://v6.ident.me",
		}
		client := &http.Client{Timeout: 5 * time.Second}
		for _, provider := range providers {
			resp, err := client.Get(provider)
			if err == nil {
				body, readErr := io.ReadAll(resp.Body)
				resp.Body.Close()
				if readErr == nil {
					ip := strings.TrimSpace(string(body))
					if len(ip) > 0 {
						publicIPv6 = ip
						return
					}
				}
			}
		}
		publicIPv6 = "N/A"
	}()
}

type ServerService struct {
	SingBoxService
}

func (s *ServerService) GetStatus(request string) *map[string]interface{} {
	status := make(map[string]interface{}, 0)
	requests := strings.Split(request, ",")
	for _, req := range requests {
		switch req {
		case "cpu":
			status["cpu"] = s.GetCpuPercent()
		case "mem":
			status["mem"] = s.GetMemInfo()
		case "disk":
			status["disk"] = s.GetDiskInfo()
		case "net":
			status["net"] = s.GetNetInfo()
		case "sys":
			status["uptime"] = s.GetUptime()
			status["sys"] = s.GetSystemInfo()
		case "sbd":
			status["sbd"] = s.GetSingboxInfo()
		}
	}
	return &status
}

func (s *ServerService) GetCpuPercent() float64 {
	percents, err := cpu.Percent(0, false)
	if err != nil {
		logger.Warning("get cpu percent failed:", err)
		return 0
	} else {
		return percents[0]
	}
}

func (s *ServerService) GetUptime() uint64 {
	upTime, err := host.Uptime()
	if err != nil {
		logger.Warning("get uptime failed:", err)
		return 0
	} else {
		return upTime
	}
}

func (s *ServerService) GetMemInfo() map[string]interface{} {
	info := make(map[string]interface{}, 0)
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		logger.Warning("get virtual memory failed:", err)
	} else {
		info["current"] = memInfo.Used
		info["total"] = memInfo.Total
	}
	return info
}

func (s *ServerService) GetDiskInfo() map[string]interface{} {
	info := make(map[string]interface{}, 0)
	diskInfo, err := disk.Usage("/")
	if err != nil {
		logger.Warning("get disk usage failed:", err)
	} else {
		info["current"] = diskInfo.Used
		info["total"] = diskInfo.Total
	}
	return info
}

func (s *ServerService) GetNetInfo() map[string]interface{} {
	info := make(map[string]interface{}, 0)
	ioStats, err := net.IOCounters(false)
	if err != nil {
		logger.Warning("get io counters failed:", err)
	} else if len(ioStats) > 0 {
		ioStat := ioStats[0]
		info["sent"] = ioStat.BytesSent
		info["recv"] = ioStat.BytesRecv
		info["psent"] = ioStat.PacketsSent
		info["precv"] = ioStat.PacketsRecv
	} else {
		logger.Warning("can not find io counters")
	}
	return info
}

func (s *ServerService) GetSingboxInfo() map[string]interface{} {
	info := make(map[string]interface{}, 0)
	sysStats, err := s.SingBoxService.GetSysStats()
	if err == nil {
		info["running"] = true
		info["stats"] = sysStats
	} else {
		info["running"] = s.SingBoxService.IsRunning()
	}
	return info
}

func (s *ServerService) GetSystemInfo() map[string]interface{} {
	info := make(map[string]interface{}, 0)
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	info["appMem"] = rtm.Sys
	info["appThreads"] = uint32(runtime.NumGoroutine())
	cpuInfo, err := cpu.Info()
	if err == nil {
		info["cpuType"] = cpuInfo[0].ModelName
	}
	info["cpuCount"] = runtime.NumCPU()
	info["hostName"], _ = os.Hostname()
	info["appVersion"] = config.GetVersion()
	ipv4 := make([]string, 0)
	ipv6 := make([]string, 0)
	// get ip address
	netInterfaces, _ := net.Interfaces()
	for i := 0; i < len(netInterfaces); i++ {
		if len(netInterfaces[i].Flags) > 2 && netInterfaces[i].Flags[0] == "up" && netInterfaces[i].Flags[1] != "loopback" {
			addrs := netInterfaces[i].Addrs

			for _, address := range addrs {
				if strings.Contains(address.Addr, ".") {
					ipv4 = append(ipv4, address.Addr)
				} else if address.Addr[0:6] != "fe80::" {
					ipv6 = append(ipv6, address.Addr)
				}
			}
		}
	}
	info["ipv4"] = ipv4
	info["ipv6"] = ipv6
	info["wanIpv4"] = publicIPv4
	info["wanIpv6"] = publicIPv6

	return info
}

func (s *ServerService) GetLogs(service string, count string, level string) []string {
	c, err := strconv.Atoi(count)
	if err != nil || c <= 0 {
		c = 100
	}
	if c > 1000 {
		c = 1000
	}

	var lines []string
	if service == "sing-box" {
		safeService := "sing-box"

		allowedLevels := map[string]bool{
			"emerg":   true,
			"alert":   true,
			"crit":    true,
			"err":     true,
			"warning": true,
			"notice":  true,
			"info":    true,
			"debug":   true,
		}
		safeLevel := "info"
		if allowedLevels[level] {
			safeLevel = level
		}

		cmdArgs := []string{"journalctl", "-u", safeService, "--no-pager", "-n", strconv.Itoa(c), "-p", safeLevel}
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			logPath := filepath.Join(config.GetBinFolderPath(), "sing-box.log")
			fallbackLines, fileErr := readPhysicalLogs(logPath, c, level)
			if fileErr != nil {
				return []string{"Failed to run journalctl command and sing-box.log reading failed: " + fileErr.Error()}
			}
			return fallbackLines
		}
		lines = strings.Split(out.String(), "\n")
	} else {
		lines = logger.GetLogs(c, level)
	}

	return lines
}

func readPhysicalLogs(logPath string, count int, level string) ([]string, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	levelPriority := map[string]int{
		"debug":   1,
		"info":    2,
		"notice":  2,
		"warning": 3,
		"warn":    3,
		"err":     4,
		"error":   4,
		"crit":    5,
		"alert":   5,
		"emerg":   5,
	}

	targetPri := levelPriority[strings.ToLower(level)]
	if targetPri == 0 {
		targetPri = 2 // default info
	}

	var matchedLines []string
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		lineUpper := strings.ToUpper(line)
		linePri := 2 // default INFO

		if strings.Contains(lineUpper, "DEBUG") {
			linePri = 1
		} else if strings.Contains(lineUpper, "INFO") {
			linePri = 2
		} else if strings.Contains(lineUpper, "WARN") {
			linePri = 3
		} else if strings.Contains(lineUpper, "ERR") || strings.Contains(lineUpper, "FATAL") {
			linePri = 4
		}

		if linePri >= targetPri {
			matchedLines = append(matchedLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		if len(matchedLines) > 0 {
			return limitLines(matchedLines, count), nil
		}
		return nil, err
	}

	return limitLines(matchedLines, count), nil
}

func limitLines(lines []string, count int) []string {
	if len(lines) <= count {
		return lines
	}
	return lines[len(lines)-count:]
}

func (s *ServerService) GenKeypair(keyType string, options string) []string {
	if keyType == "x25519" {
		keyType = "reality"
	}

	if len(keyType) == 0 {
		return []string{"No keypair to generate"}
	}
	allowedTypes := map[string]bool{
		"tls":       true,
		"ech":       true,
		"reality":   true,
		"shadowtls": true,
	}
	if !allowedTypes[keyType] {
		return []string{"Invalid keypair type"}
	}

	sbExec := s.GetBinaryPath()
	cmdArgs := []string{"generate", keyType + "-keypair"}
	if keyType == "tls" || keyType == "ech" {
		cleanedOpt := ""
		for _, r := range options {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' || r == ':' {
				cleanedOpt += string(r)
			}
		}
		if cleanedOpt == "" {
			cleanedOpt = "localhost"
		}
		cmdArgs = append(cmdArgs, cleanedOpt)
	}
	cmd := exec.Command(sbExec, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []string{"Failed to generate keypair"}
	}
	return strings.Split(out.String(), "\n")
}
