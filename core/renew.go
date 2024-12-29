package core

import (
	"cert/config"
	"encoding/json"
	"os"
	"time"
)

// 检测证书是否过期
func CheckCertExpire(cfg *config.Config) (bool, error) {
	// 当前时间是否超过json内的RenewTime
	timeNow := time.Now()
	jsonData, err := os.ReadFile(cfg.Path.Json)
	if err != nil {
		return false, err
	}
	var certInfo jsonStruct
	err = json.Unmarshal(jsonData, &certInfo)
	if err != nil {
		return false, err
	}

	renewTime, err := time.Parse(time.RFC3339, certInfo.RenewTime)
	if err != nil {
		return false, err
	}

	if timeNow.After(renewTime) {
		return true, nil
	}

	return false, nil
}

// 循环检测证书是否过期，如果过期则重新获取()
func LoopCheckCertExpire(cfg *config.Config) error {
	for {
		expire, err := CheckCertExpire(cfg)
		if err != nil {
			return err
		}
		if expire {
			err = GetNewCert(cfg)
			if err != nil {
				return err
			}
		}
		time.Sleep(time.Hour * 24)
	}
}
