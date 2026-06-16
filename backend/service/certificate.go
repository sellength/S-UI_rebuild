package service

import (
	"encoding/json"
	"fmt"
	"s-ui/database"
	"s-ui/database/model"
	"strings"
	"time"
)

type CertificateService struct {
}

const dnsCredentialKeepEncrypted = "__SUI_KEEP_ENCRYPTED__"
const certificatePEMKeepEncrypted = "__SUI_CERTIFICATE_PEM_ENCRYPTED__"

func (s *CertificateService) GetDNSProviders() ([]model.DNSProvider, error) {
	db := database.GetDB()
	providers := []model.DNSProvider{}
	err := db.Model(model.DNSProvider{}).Order("id asc").Scan(&providers).Error
	if err != nil {
		return nil, err
	}
	for i := range providers {
		if providers[i].CredentialsEncrypted == "" && hasMeaningfulJSON(providers[i].Config) {
			if err := s.encryptLegacyDNSProviderConfig(&providers[i]); err == nil {
				_ = db.Save(&providers[i]).Error
			}
		}
		if providers[i].CredentialsEncrypted != "" {
			providers[i].Config = redactedDNSProviderConfig(providers[i].Type)
			providers[i].CredentialsEncrypted = "true"
		} else if hasMeaningfulJSON(providers[i].Config) {
			providers[i].Config = nil
		}
	}
	return providers, nil
}

func (s *CertificateService) SaveDNSProvider(provider *model.DNSProvider) error {
	provider.Name = strings.TrimSpace(provider.Name)
	provider.Type = strings.TrimSpace(provider.Type)

	if provider.Name == "" {
		return fmt.Errorf("dns provider name is required")
	}
	if provider.Type == "" {
		return fmt.Errorf("dns provider type is required")
	}

	now := time.Now().Unix()
	db := database.GetDB()
	existing := model.DNSProvider{}
	if provider.Id == 0 {
		provider.CreatedAt = now
	} else {
		if err := db.Model(model.DNSProvider{}).Where("id = ?", provider.Id).First(&existing).Error; err != nil {
			return err
		}
		if provider.CreatedAt == 0 {
			provider.CreatedAt = existing.CreatedAt
		}
	}
	provider.UpdatedAt = now
	if err := s.prepareDNSProviderCredentials(provider, &existing); err != nil {
		return err
	}

	return db.Save(provider).Error
}

func (s *CertificateService) prepareDNSProviderCredentials(provider *model.DNSProvider, existing *model.DNSProvider) error {
	raw := strings.TrimSpace(string(provider.Config))
	if raw == "" {
		raw = "{}"
	}

	if provider.Id > 0 && existing != nil && existing.Id > 0 {
		if containsKeepEncryptedPlaceholder(provider.Config) {
			if existing.CredentialsEncrypted == "" && hasMeaningfulJSON(existing.Config) {
				if err := s.encryptLegacyDNSProviderConfig(existing); err != nil {
					return err
				}
			}
			if existing.CredentialsEncrypted == "" {
				return fmt.Errorf("existing dns provider has no encrypted credentials to keep")
			}
			provider.CredentialsEncrypted = existing.CredentialsEncrypted
			provider.Config = nil
			return nil
		}
		if raw == "{}" && existing.CredentialsEncrypted != "" {
			provider.CredentialsEncrypted = existing.CredentialsEncrypted
			provider.Config = nil
			return nil
		}
	}

	credentials := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &credentials); err != nil {
		return fmt.Errorf("dns provider credentials json: %w", err)
	}
	if len(credentials) == 0 {
		return fmt.Errorf("dns provider credentials are required")
	}
	if _, err := acmeDNSEnv(provider.Type, credentials); err != nil {
		return err
	}
	canonical, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("encode dns provider credentials: %w", err)
	}
	encrypted, err := encryptSecret(canonical)
	if err != nil {
		return err
	}
	provider.CredentialsEncrypted = encrypted
	provider.Config = nil
	return nil
}

func (s *CertificateService) encryptLegacyDNSProviderConfig(provider *model.DNSProvider) error {
	if !hasMeaningfulJSON(provider.Config) {
		return nil
	}
	credentials := map[string]string{}
	if err := json.Unmarshal(provider.Config, &credentials); err != nil {
		return fmt.Errorf("dns provider credentials json: %w", err)
	}
	if _, err := acmeDNSEnv(provider.Type, credentials); err != nil {
		return err
	}
	canonical, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("encode dns provider credentials: %w", err)
	}
	encrypted, err := encryptSecret(canonical)
	if err != nil {
		return err
	}
	provider.CredentialsEncrypted = encrypted
	provider.Config = nil
	return nil
}

func (s *CertificateService) DeleteDNSProvider(id uint) error {
	if id == 0 {
		return fmt.Errorf("dns provider id is required")
	}
	db := database.GetDB()
	return db.Where("id = ?", id).Delete(model.DNSProvider{}).Error
}

func (s *CertificateService) GetCertificates() ([]model.Certificate, error) {
	db := database.GetDB()
	certificates := []model.Certificate{}
	err := db.Model(model.Certificate{}).Order("id asc").Scan(&certificates).Error
	if err != nil {
		return nil, err
	}
	return certificates, nil
}

func (s *CertificateService) SaveCertificate(certificate *model.Certificate) error {
	certificate.Name = strings.TrimSpace(certificate.Name)
	certificate.Source = strings.TrimSpace(certificate.Source)

	if certificate.Name == "" {
		return fmt.Errorf("certificate name is required")
	}
	if certificate.Source == "" {
		certificate.Source = "manual"
	}

	now := time.Now().Unix()
	db := database.GetDB()
	if certificate.Id == 0 {
		certificate.CreatedAt = now
	} else if certificate.CreatedAt == 0 {
		existing := model.Certificate{}
		if err := db.Model(model.Certificate{}).Where("id = ?", certificate.Id).First(&existing).Error; err != nil {
			return err
		}
		certificate.CreatedAt = existing.CreatedAt
	}
	certificate.UpdatedAt = now

	return db.Save(certificate).Error
}

func (s *CertificateService) DeleteCertificate(id uint) error {
	if id == 0 {
		return fmt.Errorf("certificate id is required")
	}
	db := database.GetDB()
	return db.Where("id = ?", id).Delete(model.Certificate{}).Error
}

func (s *CertificateService) GetCertificateVersions(certificateId uint) ([]model.CertificateVersion, error) {
	db := database.GetDB()
	versions := []model.CertificateVersion{}
	query := db.Model(model.CertificateVersion{}).Order("id desc")
	if certificateId > 0 {
		query = query.Where("certificate_id = ?", certificateId)
	}
	err := query.Scan(&versions).Error
	if err != nil {
		return nil, err
	}
	for i := range versions {
		if err := s.encryptLegacyCertificateVersion(&versions[i]); err == nil {
			_ = db.Save(&versions[i]).Error
		}
		redactCertificateVersion(&versions[i])
	}
	return versions, nil
}

func (s *CertificateService) SaveCertificateVersion(version *model.CertificateVersion) error {
	if version.CertificateId == 0 {
		return fmt.Errorf("certificate id is required")
	}
	version.Fingerprint = strings.TrimSpace(version.Fingerprint)

	db := database.GetDB()
	existing := model.CertificateVersion{}
	if version.Id > 0 {
		if err := db.Model(model.CertificateVersion{}).Where("id = ?", version.Id).First(&existing).Error; err != nil {
			return err
		}
		if version.CreatedAt == 0 {
			version.CreatedAt = existing.CreatedAt
		}
	}
	if version.CreatedAt == 0 {
		version.CreatedAt = time.Now().Unix()
	}
	if err := s.prepareCertificateVersionSecrets(version, &existing); err != nil {
		return err
	}
	if err := db.Save(version).Error; err != nil {
		return err
	}

	certificate := model.Certificate{}
	if err := db.Model(model.Certificate{}).Where("id = ?", version.CertificateId).First(&certificate).Error; err != nil {
		return err
	}
	certificate.ActiveVersionId = version.Id
	certificate.UpdatedAt = time.Now().Unix()
	return db.Save(&certificate).Error
}

func (s *CertificateService) prepareCertificateVersionSecrets(version *model.CertificateVersion, existing *model.CertificateVersion) error {
	fullchain := strings.TrimSpace(version.FullchainPEMEncrypted)
	privateKey := strings.TrimSpace(version.PrivateKeyPEMEncrypted)

	if version.Id > 0 && existing != nil && existing.Id > 0 {
		if fullchain == "" || fullchain == certificatePEMKeepEncrypted {
			version.FullchainPEMEncrypted = existing.FullchainPEMEncrypted
		}
		if privateKey == "" || privateKey == certificatePEMKeepEncrypted {
			version.PrivateKeyPEMEncrypted = existing.PrivateKeyPEMEncrypted
		}
	}

	if version.FullchainPEMEncrypted != "" && !isEncryptedSecret(version.FullchainPEMEncrypted) {
		encrypted, err := encryptSecret([]byte(version.FullchainPEMEncrypted))
		if err != nil {
			return fmt.Errorf("encrypt certificate fullchain: %w", err)
		}
		version.FullchainPEMEncrypted = encrypted
	}
	if version.PrivateKeyPEMEncrypted != "" && !isEncryptedSecret(version.PrivateKeyPEMEncrypted) {
		encrypted, err := encryptSecret([]byte(version.PrivateKeyPEMEncrypted))
		if err != nil {
			return fmt.Errorf("encrypt certificate private key: %w", err)
		}
		version.PrivateKeyPEMEncrypted = encrypted
	}
	return nil
}

func (s *CertificateService) encryptLegacyCertificateVersion(version *model.CertificateVersion) error {
	changed := false
	if strings.TrimSpace(version.FullchainPEMEncrypted) != "" && !isEncryptedSecret(version.FullchainPEMEncrypted) {
		encrypted, err := encryptSecret([]byte(version.FullchainPEMEncrypted))
		if err != nil {
			return err
		}
		version.FullchainPEMEncrypted = encrypted
		changed = true
	}
	if strings.TrimSpace(version.PrivateKeyPEMEncrypted) != "" && !isEncryptedSecret(version.PrivateKeyPEMEncrypted) {
		encrypted, err := encryptSecret([]byte(version.PrivateKeyPEMEncrypted))
		if err != nil {
			return err
		}
		version.PrivateKeyPEMEncrypted = encrypted
		changed = true
	}
	if !changed {
		return nil
	}
	return nil
}

func redactCertificateVersion(version *model.CertificateVersion) {
	if strings.TrimSpace(version.FullchainPEMEncrypted) != "" {
		version.FullchainPEMEncrypted = certificatePEMKeepEncrypted
	}
	if strings.TrimSpace(version.PrivateKeyPEMEncrypted) != "" {
		version.PrivateKeyPEMEncrypted = certificatePEMKeepEncrypted
	}
}

func certificateVersionPEM(version model.CertificateVersion) (string, string, error) {
	fullchain, err := decryptMaybeEncryptedPEM(version.FullchainPEMEncrypted)
	if err != nil {
		return "", "", fmt.Errorf("decrypt certificate fullchain: %w", err)
	}
	privateKey, err := decryptMaybeEncryptedPEM(version.PrivateKeyPEMEncrypted)
	if err != nil {
		return "", "", fmt.Errorf("decrypt certificate private key: %w", err)
	}
	return fullchain, privateKey, nil
}

func decryptMaybeEncryptedPEM(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !isEncryptedSecret(value) {
		return value, nil
	}
	plain, err := decryptSecret(value)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
