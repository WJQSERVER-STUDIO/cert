package core

import (
	"cert/config"
	"cert/logger"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

// json 结构体
type jsonStruct struct {
	Version            int
	SerialNumber       string
	SignatureAlgorithm string
	Issuer             string
	Validity           struct {
		NotBefore string
		NotAfter  string
	}
	Subject   string
	RenewTime string
}

// 预检测证书文件是否存在
func PreCheck(cfg *config.Config) (bool, error) {
	_, err := os.Stat(cfg.Path.Json)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func GetNewCert(cfg *config.Config) error {

	// 检测证书文件是否存在
	exist, err := PreCheck(cfg)
	if err != nil {
		return err
	}

	if exist {
		// 检测证书是否过期
		expire, err := CheckCertExpire(cfg)
		if err != nil {
			return err
		}

		if expire {
			logInfo("证书未到期，无需续签")
			return nil
		} else {
			logInfo("证书已到期，准备续签")
		}
	} else {
		logInfo("证书文件不存在，准备获取新证书")
	}

	// Cloudflare API token
	os.Setenv("CLOUDFLARE_DNS_API_TOKEN", cfg.Account.Token)

	// 设置 DNS 解析器为 1.1.1.1
	os.Setenv("LEGO_DNS_RESOLVERS", "1.1.1.1")

	// 创建用户
	user, err := NewUser(cfg.Account.Email)
	if err != nil {
		return err
	}

	// 配置lego客户端
	config := lego.NewConfig(user)
	config.Certificate.KeyType = certcrypto.RSA2048

	// 创建lego客户端
	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}

	// 使用Cloudflare DNS挑战
	provider, err := cloudflare.NewDNSProvider()
	if err != nil {
		return err
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return err
	}

	// 注册用户
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	user.Registration = reg

	// 请求证书
	request := certificate.ObtainRequest{
		Domains: []string{cfg.Domain.Name},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	// 保存证书
	err = saveCert(cfg, certificates)
	if err != nil {
		return err
	}

	// 保存证书信息到JSON
	err = saveCertInfoAsJSON(cfg, certificates.Certificate)
	if err != nil {
		return err
	}

	log.Println("证书获取成功")
	return nil
}

// NewUser 创建一个新的用户
func NewUser(email string) (*User, error) {
	key, err := generatePrivateKey()
	if err != nil {
		return nil, err
	}

	return &User{
		Email: email,
		Key:   key,
	}, err
}

// User 用户结构体
type User struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

// GetEmail 获取用户电子邮件
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration 获取用户注册信息
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey 获取用户私钥
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

// generatePrivateKey 生成私钥
func generatePrivateKey() (crypto.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// saveCert 保存证书
func saveCert(cfg *config.Config, certificates *certificate.Resource) error {
	err := os.WriteFile(cfg.Path.Cert, certificates.Certificate, 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.Path.Key, certificates.PrivateKey, 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.Path.CaCert, certificates.IssuerCertificate, 0600)
	if err != nil {
		return err
	}

	return nil
}

// saveCertInfoAsJSON 将证书信息保存为JSON文件
func saveCertInfoAsJSON(cfg *config.Config, certPEM []byte) error {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	// 设置续签时间为证书到期前30天
	renewTime := cert.NotAfter.AddDate(0, 0, -30)

	/*certInfo := map[string]interface{}{
		"Version":            cert.Version,
		"SerialNumber":       cert.SerialNumber.String(),
		"SignatureAlgorithm": cert.SignatureAlgorithm.String(),
		"Issuer":             cert.Issuer.String(),
		"Validity": map[string]string{
			"NotBefore": cert.NotBefore.Format(time.RFC3339),
			"NotAfter":  cert.NotAfter.Format(time.RFC3339),
		},
		"Subject":   cert.Subject.String(),
		"RenewTime": renewTime.Format(time.RFC3339),
	}*/

	certInfo := jsonStruct{
		Version:            cert.Version,
		SerialNumber:       cert.SerialNumber.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		Issuer:             cert.Issuer.String(),
		Validity: struct {
			NotBefore string
			NotAfter  string
		}{
			NotBefore: cert.NotBefore.Format(time.RFC3339),
			NotAfter:  cert.NotAfter.Format(time.RFC3339),
		},
		Subject:   cert.Subject.String(),
		RenewTime: renewTime.Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(certInfo, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.Path.Json, data, 0600)
	if err != nil {
		return err
	}

	return nil
}
