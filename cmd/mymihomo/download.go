package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultFetchTimeoutSec = 15
	defaultFetchRetry      = 2
)

func downloadConfig(confFile string) error {
	confURL := strings.TrimSpace(os.Getenv("CONF_URL"))
	if confURL == "" || confURL == "http://test.com" {
		return fmt.Errorf("错误: CONF_URL 环境变量未设置或使用了默认值\n\n请在 docker-compose.yml 中配置你的订阅地址:\n  environment:\n    - CONF_URL=https://your-subscription-url")
	}

	// 创建目录
	confPath := filepath.Dir(confFile)
	if err := os.MkdirAll(confPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	fmt.Printf("从 %s 下载配置文件并写入到 %s\n", confURL, confFile)

	// 下载配置文件
	data, err := fetchConfig(confURL)
	if err != nil {
		return fmt.Errorf("配置文件下载失败: %w", err)
	}

	// Base64 转换
	if getEnvBool("BASE64_CONVERT", false) {
		fmt.Println("配置文件转换中...")
		decoded, err := decodeBase64(data)
		if err != nil {
			return fmt.Errorf("base64 解码失败: %w", err)
		}
		data = decoded
	}

	// 解析 YAML
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("YAML 解析失败: %w", err)
	}

	// 修改配置
	modifyConfig(config)

	// 合并自定义规则
	customConf := os.Getenv("CUSTOM_CONF")
	if customConf != "" {
		if err := mergeCustomRules(config, customConf); err != nil {
			fmt.Printf("警告: 合并自定义规则失败: %v\n", err)
		}
	}

	// 写入文件
	output, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("YAML 序列化失败: %w", err)
	}

	if err := os.WriteFile(confFile, output, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	// 渲染导航页
	indexPath := "/root/.config/mihomo/ui/index.html"
	if err := renderIndex(indexPath); err != nil {
		fmt.Printf("警告: 渲染导航页失败: %v\n", err)
	}

	fmt.Println("配置文件处理完成")
	return nil
}

func fetchConfig(url string) ([]byte, error) {
	if strings.HasPrefix(url, "file://") {
		return os.ReadFile(strings.TrimPrefix(url, "file://"))
	}

	timeoutSec := getEnvInt("CONF_TIMEOUT_SEC", defaultFetchTimeoutSec)
	retryCount := getEnvInt("CONF_RETRY", defaultFetchRetry)
	if retryCount < 0 {
		retryCount = defaultFetchRetry
	}
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	userAgent := getEnvDefault("CLASH_SUB_UA", "Wget/1.21.3")

	var lastErr error
	for attempt := 0; attempt <= retryCount; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = readErr
			} else if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("HTTP 状态码: %d, 响应: %.200s", resp.StatusCode, body)
			} else {
				return body, nil
			}
		}

		if attempt < retryCount {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if lastErr == nil {
		lastErr = errors.New("未知下载错误")
	}
	return nil, lastErr
}

func decodeBase64(data []byte) ([]byte, error) {
	content := strings.TrimSpace(string(data))
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil {
		return decoded, nil
	}

	rawDecoded, rawErr := base64.RawStdEncoding.DecodeString(content)
	if rawErr == nil {
		return rawDecoded, nil
	}

	return nil, fmt.Errorf("标准解码失败: %v; Raw 解码失败: %v", err, rawErr)
}

func modifyConfig(config map[string]interface{}) {
	externalBind := getEnvDefault("EXTERNAL_BIND", "0.0.0.0")
	externalPort := getEnvDefault("EXTERNAL_PORT", "9090")

	// 设置 external-controller
	config["external-controller"] = fmt.Sprintf("%s:%s", externalBind, externalPort)
	config["external-ui"] = "/root/.config/mihomo/ui"

	// 设置 secret
	if secret := os.Getenv("EXTERNAL_SECRET"); secret != "" {
		config["secret"] = secret
	}

	// 开启局域网连接
	config["allow-lan"] = true

	// 删除旧配置，使用环境变量重新设置
	delete(config, "port")
	delete(config, "socks-port")
	delete(config, "mixed-port")

	// HTTP 代理端口
	if httpPort := os.Getenv("HTTP_PORT"); httpPort != "" {
		if port, err := parsePort(httpPort); err == nil {
			config["port"] = port
			fmt.Printf("设置 HTTP 代理端口: %d\n", port)
		} else {
			fmt.Printf("警告: HTTP_PORT 无效: %v\n", err)
		}
	}

	// SOCKS5 代理端口
	if socksPort := os.Getenv("SOCKS_PORT"); socksPort != "" {
		if port, err := parsePort(socksPort); err == nil {
			config["socks-port"] = port
			fmt.Printf("设置 SOCKS5 代理端口: %d\n", port)
		} else {
			fmt.Printf("警告: SOCKS_PORT 无效: %v\n", err)
		}
	}

	// Mixed 代理端口 (HTTP + SOCKS5)
	if mixedPort := os.Getenv("MIXED_PORT"); mixedPort != "" {
		if port, err := parsePort(mixedPort); err == nil {
			config["mixed-port"] = port
			fmt.Printf("设置 Mixed 代理端口: %d\n", port)
		} else {
			fmt.Printf("警告: MIXED_PORT 无效: %v\n", err)
		}
	}

	_, hasHTTPEnv := os.LookupEnv("HTTP_PORT")
	_, hasSOCKSEnv := os.LookupEnv("SOCKS_PORT")
	_, hasMixedEnv := os.LookupEnv("MIXED_PORT")
	if !hasHTTPEnv && !hasSOCKSEnv && !hasMixedEnv {
		config["mixed-port"] = 7890
		fmt.Println("未显式配置代理端口，回退设置 Mixed 端口: 7890")
	}

	// TUN 模式配置
	if tunEnable := os.Getenv("TUN_ENABLE"); tunEnable == "true" {
		tunConfig := map[string]interface{}{
			"enable":                true,
			"stack":                 getEnvDefault("TUN_STACK", "system"),
			"auto-route":            getEnvBool("TUN_AUTO_ROUTE", true),
			"auto-detect-interface": getEnvBool("TUN_AUTO_DETECT", true),
			"dns-hijack":            []string{"any:53"},
		}
		config["tun"] = tunConfig
		fmt.Println("启用 TUN 模式")
	}

	fmt.Println("修改配置文件中的ui界面指向路径")
}

func mergeCustomRules(config map[string]interface{}, customConfPath string) error {
	if _, err := os.Stat(customConfPath); os.IsNotExist(err) {
		return nil
	}

	fmt.Printf("合并自定义规则文件... %s\n", customConfPath)

	data, err := os.ReadFile(customConfPath)
	if err != nil {
		return err
	}

	var customConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &customConfig); err != nil {
		return err
	}

	// 合并 rules
	if customRules, ok := customConfig["rules"].([]interface{}); ok {
		existingRules, _ := config["rules"].([]interface{})
		config["rules"] = append(existingRules, customRules...)
	}

	return nil
}

func getEnvDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	val = strings.ToLower(strings.TrimSpace(val))
	return val == "true" || val == "1"
}

func getEnvInt(key string, defaultVal int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("警告: %s=%q 不是合法整数，使用默认值: %d\n", key, val, defaultVal)
		return defaultVal
	}
	return num
}

func parsePort(s string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("端口格式错误: %w", err)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("端口范围必须在 1-65535，当前: %d", port)
	}
	return port, nil
}
