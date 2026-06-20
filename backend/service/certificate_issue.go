package service

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"s-ui/config"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

type commandRunner func(name string, args []string, env []string) ([]byte, error)

func defaultCommandRunner(name string, args []string, env []string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	return cmd.CombinedOutput()
}

func (s *CertificateService) IssueCertificate(certificateId uint) (*model.CertificateVersion, error) {
	return s.issueCertificateWithRunner(certificateId, defaultCommandRunner)
}

func (s *CertificateService) issueCertificateWithRunner(certificateId uint, runner commandRunner) (*model.CertificateVersion, error) {
	if certificateId == 0 {
		return nil, fmt.Errorf("certificate id is required")
	}

	db := database.GetDB()
	certificate := model.Certificate{}
	if err := db.Model(model.Certificate{}).Where("id = ?", certificateId).First(&certificate).Error; err != nil {
		return nil, err
	}
	if certificate.Source != "acme-dns01" && certificate.Source != "acme.sh" {
		return nil, fmt.Errorf("certificate source %q does not support automatic issue", certificate.Source)
	}

	domains, err := certificateDomains(certificate.Domains)
	if err != nil {
		return nil, err
	}

	certConfig := map[string]interface{}{}
	if len(certificate.Config) > 0 {
		if err := json.Unmarshal(certificate.Config, &certConfig); err != nil {
			return nil, fmt.Errorf("certificate config json: %w", err)
		}
	}

	dnsAPI := configString(certConfig, "acmeDnsApi")
	env := []string{}
	if certificate.Source == "acme-dns01" {
		provider, providerEnv, err := s.acmeDNSProvider(certificate.DNSProviderId)
		if err != nil {
			return nil, err
		}
		dnsAPI = acmeDNSAPI(provider.Type)
		env = providerEnv
	}
	if dnsAPI == "" {
		return nil, fmt.Errorf("acme dns api is required")
	}

	workDir := filepath.Join(config.GetCertificateWorkDir(), fmt.Sprintf("certificate-%d", certificate.Id))
	if err := os.MkdirAll(workDir, 0700); err != nil {
		return nil, err
	}
	fullchainPath := filepath.Join(workDir, "fullchain.pem")
	keyPath := filepath.Join(workDir, "privkey.pem")

	acmeHome := configString(certConfig, "acmeHome")
	if acmeHome == "" {
		acmeHome = config.GetACMEHome()
	}
	mainDomain := primaryDomain(domains)
	server := acmeServer(configString(certConfig, "ca"))

	issueArgs := []string{"--issue", "--force", "--dns", dnsAPI, "--server", server, "--home", acmeHome}
	for _, domain := range domains {
		issueArgs = append(issueArgs, "-d", domain)
	}
	if output, err := runner(config.GetACMEShPath(), issueArgs, env); err != nil {
		return nil, fmt.Errorf("acme.sh issue failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	installArgs := []string{"--install-cert", "-d", mainDomain, "--home", acmeHome, "--key-file", keyPath, "--fullchain-file", fullchainPath}
	if output, err := runner(config.GetACMEShPath(), installArgs, env); err != nil {
		return nil, fmt.Errorf("acme.sh install-cert failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	fullchain, err := os.ReadFile(fullchainPath)
	if err != nil {
		return nil, err
	}
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	fingerprint, notBefore, notAfter := certificateMetadata(fullchain)

	version := &model.CertificateVersion{
		CertificateId:          certificate.Id,
		Fingerprint:            fingerprint,
		FullchainPEMEncrypted:  string(fullchain),
		PrivateKeyPEMEncrypted: string(privateKey),
		NotBefore:              notBefore,
		NotAfter:               notAfter,
		CreatedAt:              time.Now().Unix(),
	}
	if err := s.SaveCertificateVersion(version); err != nil {
		return nil, err
	}
	return version, nil
}

func (s *CertificateService) acmeDNSProvider(providerId uint) (model.DNSProvider, []string, error) {
	if providerId == 0 {
		return model.DNSProvider{}, nil, fmt.Errorf("dns provider is required")
	}
	db := database.GetDB()
	provider := model.DNSProvider{}
	if err := db.Model(model.DNSProvider{}).Where("id = ? AND enable = ?", providerId, true).First(&provider).Error; err != nil {
		if database.IsNotFound(err) {
			return model.DNSProvider{}, nil, fmt.Errorf("DNS provider is disabled or not found")
		}
		return model.DNSProvider{}, nil, err
	}
	credentials, err := dnsProviderCredentials(provider)
	if err != nil {
		return model.DNSProvider{}, nil, err
	}
	env, err := acmeDNSEnv(provider.Type, credentials)
	if err != nil {
		return model.DNSProvider{}, nil, err
	}
	return provider, env, nil
}

func dnsProviderCredentials(provider model.DNSProvider) (map[string]string, error) {
	raw := []byte(strings.TrimSpace(string(provider.Config)))
	if provider.CredentialsEncrypted != "" {
		decrypted, err := decryptSecret(provider.CredentialsEncrypted)
		if err != nil {
			return nil, err
		}
		raw = decrypted
	}
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	credentials := map[string]string{}
	if err := json.Unmarshal(raw, &credentials); err != nil {
		return nil, fmt.Errorf("dns provider credentials json: %w", err)
	}
	return credentials, nil
}

func certificateDomains(raw json.RawMessage) ([]string, error) {
	domains := []string{}
	if err := json.Unmarshal(raw, &domains); err != nil {
		return nil, fmt.Errorf("certificate domains json: %w", err)
	}
	filtered := []string{}
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			filtered = append(filtered, domain)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("at least one certificate domain is required")
	}
	return filtered, nil
}

func acmeDNSAPI(providerType string) string {
	switch strings.ToLower(strings.TrimSpace(providerType)) {
	case "cloudflare":
		return "dns_cf"
	case "dnspod":
		return "dns_dp"
	case "aliyun":
		return "dns_ali"
	case "route53":
		return "dns_aws"
	default:
		return ""
	}
}

func acmeDNSEnv(providerType string, credentials map[string]string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(providerType)) {
	case "cloudflare":
		token := firstCredential(credentials, "apiToken", "token", "CF_Token")
		if token == "" {
			return nil, fmt.Errorf("cloudflare CF_Token is required")
		}
		env := []string{"CF_Token=" + token}
		if zoneID := firstCredential(credentials, "zoneId", "CF_Zone_ID"); zoneID != "" {
			env = append(env, "CF_Zone_ID="+zoneID)
		}
		if accountID := firstCredential(credentials, "accountId", "CF_Account_ID"); accountID != "" {
			env = append(env, "CF_Account_ID="+accountID)
		}
		return env, nil
	case "aliyun":
		key := firstCredential(credentials, "Ali_Key", "aliKey", "accessKeyId", "accessKeyID")
		secret := firstCredential(credentials, "Ali_Secret", "aliSecret", "accessKeySecret")
		if key == "" || secret == "" {
			return nil, fmt.Errorf("alidns Ali_Key and Ali_Secret are required")
		}
		return []string{"Ali_Key=" + key, "Ali_Secret=" + secret}, nil
	case "route53":
		key := firstCredential(credentials, "AWS_ACCESS_KEY_ID", "awsAccessKeyId", "accessKeyId")
		secret := firstCredential(credentials, "AWS_SECRET_ACCESS_KEY", "awsSecretAccessKey", "secretAccessKey")
		if key == "" || secret == "" {
			return nil, fmt.Errorf("route53 AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are required")
		}
		env := []string{"AWS_ACCESS_KEY_ID=" + key, "AWS_SECRET_ACCESS_KEY=" + secret}
		if region := firstCredential(credentials, "AWS_REGION", "awsRegion"); region != "" {
			env = append(env, "AWS_REGION="+region)
		}
		return env, nil
	case "dnspod":
		id := firstCredential(credentials, "DP_Id", "dpId", "id")
		key := firstCredential(credentials, "DP_Key", "dpKey", "key")
		if id == "" || key == "" {
			return nil, fmt.Errorf("dnspod DP_Id and DP_Key are required")
		}
		return []string{"DP_Id=" + id, "DP_Key=" + key}, nil
	default:
		return nil, fmt.Errorf("dns provider %q is not supported yet", providerType)
	}
}

func firstCredential(credentials map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(credentials[key]); value != "" {
			return value
		}
	}
	return ""
}

func configString(values map[string]interface{}, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func primaryDomain(domains []string) string {
	for _, domain := range domains {
		if !strings.HasPrefix(domain, "*.") {
			return domain
		}
	}
	return domains[0]
}

func acmeServer(ca string) string {
	switch strings.ToLower(strings.TrimSpace(ca)) {
	case "zerossl":
		return "zerossl"
	default:
		return "letsencrypt"
	}
}

func certificateMetadata(fullchain []byte) (string, int64, int64) {
	block, _ := pem.Decode(fullchain)
	if block == nil {
		sum := sha256.Sum256(fullchain)
		return strings.ToUpper(hex.EncodeToString(sum[:])), 0, 0
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		sum := sha256.Sum256(block.Bytes)
		return strings.ToUpper(hex.EncodeToString(sum[:])), 0, 0
	}
	sum := sha256.Sum256(cert.Raw)
	return strings.ToUpper(hex.EncodeToString(sum[:])), cert.NotBefore.Unix(), cert.NotAfter.Unix()
}
