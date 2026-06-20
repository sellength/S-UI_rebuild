package cronjob

import (
	"s-ui/database"
	"s-ui/database/model"
	"s-ui/logger"
	"s-ui/service"
	"time"
)

type CertRenewJob struct {
	service.CertificateService
}

func NewCertRenewJob() *CertRenewJob {
	return &CertRenewJob{}
}

func (j *CertRenewJob) Run() {
	db := database.GetDB()
	certs := []model.Certificate{}
	
	// 只查询开启了自动续期，并且已经有激活版本的证书
	err := db.Model(&model.Certificate{}).
		Where("auto_renew = ? AND active_version_id > 0", true).
		Scan(&certs).Error
	if err != nil {
		logger.Warning("Cronjob query auto renew certificates failed: ", err)
		return
	}

	now := time.Now().Unix()
	renewCount := 0
	
	for _, cert := range certs {
		version := model.CertificateVersion{}
		if err := db.Model(model.CertificateVersion{}).Where("id = ?", cert.ActiveVersionId).First(&version).Error; err != nil {
			logger.Warningf("Cronjob check cert %q (ID: %d) active version failed: %v", cert.Name, cert.Id, err)
			continue
		}

		// 检查是否到期前 3 天 (3 * 24 * 3600 秒)
		remainingSec := version.NotAfter - now
		if remainingSec <= 3*24*3600 {
			logger.Infof("Certificate %q (ID: %d) is expiring in %.2f days, triggering auto renewal...", cert.Name, cert.Id, float64(remainingSec)/86400.0)
			
			_, issueErr := j.CertificateService.IssueCertificate(cert.Id)
			if issueErr != nil {
				logger.Warningf("Auto renewal of certificate %q failed: %v", cert.Name, issueErr)
			} else {
				logger.Infof("Auto renewal of certificate %q completed successfully", cert.Name)
				renewCount++
			}
		}
	}

	if renewCount > 0 {
		logger.Infof("Auto cert renewal cronjob finished. Renewed %d certificates.", renewCount)
	}
}
