/*
These codes are OpenSource By WSL (WJQserver Studio License) https://raw.github.com/WJQSERVER-STUDIO/LICENSE/main/LICENSE
This part write by wjqserver
*/

package main

import (
	"cert/config"
	"cert/core"
	"flag"
	"fmt"
	"log"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"

	"github.com/gin-gonic/gin"
)

var (
	cfg        *config.Config
	configfile = "/data/cert/config/config.toml"
	router     *gin.Engine
)

// 日志模块
var (
	logw       = logger.Logw
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func ReadFlag() {
	cfgfile := flag.String("cfg", configfile, "config file path")
	flag.Parse()
	configfile = *cfgfile
}

func loadConfig() {
	var err error
	// 初始化配置
	cfg, err = config.LoadConfig(configfile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %v\n", cfg)
}

func setupLogger() {
	// 初始化日志模块
	var err error
	err = logger.Init(cfg.Log.LogFilePath, cfg.Log.MaxLogSize) // 传递日志文件路径
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logw("Logger initialized")
	logw("Init Completed")
}

// 协程启动过期检测
func setupExpireCheck(cfg *config.Config) {
	go func() {
		err := core.LoopCheckCertExpire(cfg)
		if err != nil {
			logError("Failed to check cert expire: %v\n", err)
		}
	}()
}

func init() {
	ReadFlag()
	loadConfig()
	setupLogger()
}

func main() {
	err := core.GetNewCert(cfg)
	if err != nil {
		logError("Failed to get new cert: %v\n", err)
	}
	defer logger.Close() // 确保在退出时关闭日志文件
}
