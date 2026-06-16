package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"testing"
	"time"
)

func initTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "s-ui-service-test.db")
	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
}

func TestCertificateServiceSaveCertificateDefaultsSource(t *testing.T) {
	initTestDB(t)

	service := CertificateService{}
	certificate := &model.Certificate{
		Name: "Wildcard example.com",
	}

	if err := service.SaveCertificate(certificate); err != nil {
		t.Fatalf("SaveCertificate() error = %v", err)
	}
	if certificate.Source != "manual" {
		t.Fatalf("expected default source manual, got %q", certificate.Source)
	}
	if certificate.CreatedAt == 0 || certificate.UpdatedAt == 0 {
		t.Fatalf("expected timestamps to be set")
	}
}

func TestCertificateServiceRequiresDNSProviderType(t *testing.T) {
	initTestDB(t)

	service := CertificateService{}
	err := service.SaveDNSProvider(&model.DNSProvider{Name: "Cloudflare"})
	if err == nil {
		t.Fatalf("expected missing provider type error")
	}
}

func TestCertificateServiceSaveVersionActivatesCertificate(t *testing.T) {
	initTestDB(t)
	t.Setenv("SUI_SECRET_KEY", "test-secret-key-at-least-32-characters")

	service := CertificateService{}
	certificate := &model.Certificate{Name: "Wildcard", Source: "manual"}
	if err := service.SaveCertificate(certificate); err != nil {
		t.Fatalf("SaveCertificate() error = %v", err)
	}

	version := &model.CertificateVersion{
		CertificateId:          certificate.Id,
		Fingerprint:            "abc123",
		FullchainPEMEncrypted:  "fullchain",
		PrivateKeyPEMEncrypted: "privkey",
	}
	if err := service.SaveCertificateVersion(version); err != nil {
		t.Fatalf("SaveCertificateVersion() error = %v", err)
	}

	db := database.GetDB()
	loaded := model.Certificate{}
	if err := db.Model(model.Certificate{}).Where("id = ?", certificate.Id).First(&loaded).Error; err != nil {
		t.Fatalf("load certificate: %v", err)
	}
	if loaded.ActiveVersionId != version.Id {
		t.Fatalf("expected active version %d, got %d", version.Id, loaded.ActiveVersionId)
	}
	loadedVersion := model.CertificateVersion{}
	if err := database.GetDB().Model(model.CertificateVersion{}).Where("id = ?", version.Id).First(&loadedVersion).Error; err != nil {
		t.Fatalf("load certificate version: %v", err)
	}
	if !isEncryptedSecret(loadedVersion.FullchainPEMEncrypted) || !isEncryptedSecret(loadedVersion.PrivateKeyPEMEncrypted) {
		t.Fatalf("expected encrypted certificate pem fields, got %#v", loadedVersion)
	}
	fullchain, privateKey, err := certificateVersionPEM(loadedVersion)
	if err != nil {
		t.Fatalf("certificateVersionPEM() error = %v", err)
	}
	if fullchain != "fullchain" || privateKey != "privkey" {
		t.Fatalf("unexpected decrypted certificate pem: %q %q", fullchain, privateKey)
	}
	versions, err := service.GetCertificateVersions(certificate.Id)
	if err != nil {
		t.Fatalf("GetCertificateVersions() error = %v", err)
	}
	if len(versions) != 1 || strings.Contains(versions[0].PrivateKeyPEMEncrypted, "privkey") || versions[0].PrivateKeyPEMEncrypted != certificatePEMKeepEncrypted {
		t.Fatalf("expected redacted certificate version, got %#v", versions)
	}
}

func TestCertificateServiceIssueCertificateWithCloudflareProvider(t *testing.T) {
	initTestDB(t)
	t.Setenv("SUI_CERTIFICATE_WORK_DIR", t.TempDir())
	t.Setenv("SUI_SECRET_KEY", "test-secret-key-at-least-32-characters")

	service := CertificateService{}
	provider := &model.DNSProvider{
		Enable: true,
		Name:   "cf-main",
		Type:   "cloudflare",
		Config: json.RawMessage(`{"apiToken":"test-token"}`),
	}
	if err := service.SaveDNSProvider(provider); err != nil {
		t.Fatalf("SaveDNSProvider() error = %v", err)
	}
	certificate := &model.Certificate{
		Enable:        true,
		Name:          "Wildcard",
		Source:        "acme-dns01",
		Domains:       json.RawMessage(`["*.example.com","example.com"]`),
		Config:        json.RawMessage(`{"ca":"letsencrypt"}`),
		Wildcard:      true,
		AutoRenew:     true,
		DNSProviderId: provider.Id,
	}
	if err := service.SaveCertificate(certificate); err != nil {
		t.Fatalf("SaveCertificate() error = %v", err)
	}

	certPEM, keyPEM := testCertificatePEM(t)
	runner := func(name string, args []string, env []string) ([]byte, error) {
		if len(args) > 0 && args[0] == "--install-cert" {
			writeFlagFile(t, args, "--fullchain-file", certPEM)
			writeFlagFile(t, args, "--key-file", keyPEM)
		}
		if !contains(env, "CF_Token=test-token") {
			t.Fatalf("expected cloudflare token env, got %#v", env)
		}
		return []byte("ok"), nil
	}

	version, err := service.issueCertificateWithRunner(certificate.Id, runner)
	if err != nil {
		t.Fatalf("IssueCertificate() error = %v", err)
	}
	if version.Id == 0 || version.Fingerprint == "" || version.NotAfter == 0 {
		t.Fatalf("expected saved active certificate version, got %#v", version)
	}
	if !isEncryptedSecret(version.FullchainPEMEncrypted) || !isEncryptedSecret(version.PrivateKeyPEMEncrypted) {
		t.Fatalf("expected issued certificate pem fields to be encrypted, got %#v", version)
	}
	fullchain, privateKey, err := certificateVersionPEM(*version)
	if err != nil {
		t.Fatalf("certificateVersionPEM() error = %v", err)
	}
	if !strings.Contains(fullchain, "BEGIN CERTIFICATE") || !strings.Contains(privateKey, "BEGIN RSA PRIVATE KEY") {
		t.Fatalf("unexpected decrypted issued certificate")
	}
}

func TestCertificateServiceEncryptsDNSProviderCredentials(t *testing.T) {
	initTestDB(t)
	t.Setenv("SUI_SECRET_KEY", "test-secret-key-at-least-32-characters")

	service := CertificateService{}
	provider := &model.DNSProvider{
		Enable: true,
		Name:   "cf-main",
		Type:   "cloudflare",
		Config: json.RawMessage(`{"CF_Token":"cf-token"}`),
	}
	if err := service.SaveDNSProvider(provider); err != nil {
		t.Fatalf("SaveDNSProvider() error = %v", err)
	}
	if provider.CredentialsEncrypted == "" || !isEncryptedSecret(provider.CredentialsEncrypted) {
		t.Fatalf("expected encrypted credentials, got %q", provider.CredentialsEncrypted)
	}
	if string(provider.Config) != "" {
		t.Fatalf("expected plaintext config to be cleared, got %s", string(provider.Config))
	}

	loaded := model.DNSProvider{}
	if err := database.GetDB().Model(model.DNSProvider{}).Where("id = ?", provider.Id).First(&loaded).Error; err != nil {
		t.Fatalf("load provider error = %v", err)
	}
	if strings.Contains(loaded.CredentialsEncrypted, "cf-token") || strings.Contains(string(loaded.Config), "cf-token") {
		t.Fatalf("stored provider leaked plaintext credential: %#v", loaded)
	}

	providers, err := service.GetDNSProviders()
	if err != nil {
		t.Fatalf("GetDNSProviders() error = %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected one provider, got %d", len(providers))
	}
	if strings.Contains(string(providers[0].Config), "cf-token") {
		t.Fatalf("listed provider leaked plaintext credential: %s", string(providers[0].Config))
	}
	if !strings.Contains(string(providers[0].Config), dnsCredentialKeepEncrypted) {
		t.Fatalf("expected listed provider to include keep-encrypted placeholder, got %s", string(providers[0].Config))
	}

	providers[0].Name = "cf-renamed"
	if err := service.SaveDNSProvider(&providers[0]); err != nil {
		t.Fatalf("SaveDNSProvider() with placeholder error = %v", err)
	}
	reloaded := model.DNSProvider{}
	if err := database.GetDB().Model(model.DNSProvider{}).Where("id = ?", provider.Id).First(&reloaded).Error; err != nil {
		t.Fatalf("reload provider error = %v", err)
	}
	if reloaded.Name != "cf-renamed" {
		t.Fatalf("expected renamed provider, got %q", reloaded.Name)
	}
	credentials, err := dnsProviderCredentials(reloaded)
	if err != nil {
		t.Fatalf("dnsProviderCredentials() error = %v", err)
	}
	if credentials["CF_Token"] != "cf-token" {
		t.Fatalf("expected preserved token, got %#v", credentials)
	}
}

func TestACMEDNSEnvProviders(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		credentials map[string]string
		want        []string
	}{
		{
			name:     "cloudflare",
			provider: "cloudflare",
			credentials: map[string]string{
				"CF_Token":      "cf-token",
				"CF_Zone_ID":    "zone-id",
				"CF_Account_ID": "account-id",
			},
			want: []string{"CF_Token=cf-token", "CF_Zone_ID=zone-id", "CF_Account_ID=account-id"},
		},
		{
			name:     "aliyun",
			provider: "aliyun",
			credentials: map[string]string{
				"Ali_Key":    "ali-key",
				"Ali_Secret": "ali-secret",
			},
			want: []string{"Ali_Key=ali-key", "Ali_Secret=ali-secret"},
		},
		{
			name:     "route53",
			provider: "route53",
			credentials: map[string]string{
				"AWS_ACCESS_KEY_ID":     "aws-key",
				"AWS_SECRET_ACCESS_KEY": "aws-secret",
			},
			want: []string{"AWS_ACCESS_KEY_ID=aws-key", "AWS_SECRET_ACCESS_KEY=aws-secret"},
		},
		{
			name:     "dnspod",
			provider: "dnspod",
			credentials: map[string]string{
				"DP_Id":  "dp-id",
				"DP_Key": "dp-key",
			},
			want: []string{"DP_Id=dp-id", "DP_Key=dp-key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := acmeDNSEnv(tt.provider, tt.credentials)
			if err != nil {
				t.Fatalf("acmeDNSEnv() error = %v", err)
			}
			for _, want := range tt.want {
				if !contains(env, want) {
					t.Fatalf("expected %q in env %#v", want, env)
				}
			}
		})
	}
}

func writeFlagFile(t *testing.T, args []string, flag string, content []byte) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			if err := os.WriteFile(args[i+1], content, 0600); err != nil {
				t.Fatalf("write %s: %v", flag, err)
			}
			return
		}
	}
	t.Fatalf("missing %s in args %#v", flag, args)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func testCertificatePEM(t *testing.T) ([]byte, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"example.com", "*.example.com"},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM
}
