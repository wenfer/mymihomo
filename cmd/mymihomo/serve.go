package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var configurableKeys = []string{
	"CONF_URL",
	"EXTERNAL_BIND",
	"EXTERNAL_PORT",
	"EXTERNAL_SECRET",
	"HTTP_PORT",
	"SOCKS_PORT",
	"MIXED_PORT",
	"TUN_ENABLE",
	"TUN_STACK",
	"TUN_AUTO_ROUTE",
	"TUN_AUTO_DETECT",
	"BASE64_CONVERT",
	"CUSTOM_CONF",
	"CRON_EXPRESSION",
	"CONF_TIMEOUT_SEC",
	"CONF_RETRY",
	"CLASH_SUB_UA",
}

var defaultConfigValues = map[string]string{
	"EXTERNAL_BIND":    "0.0.0.0",
	"EXTERNAL_PORT":    "9090",
	"MIXED_PORT":       "7890",
	"TUN_ENABLE":       "false",
	"TUN_STACK":        "system",
	"TUN_AUTO_ROUTE":   "true",
	"TUN_AUTO_DETECT":  "true",
	"BASE64_CONVERT":   "false",
	"CRON_EXPRESSION":  "1 * * * *",
	"CONF_TIMEOUT_SEC": "15",
	"CONF_RETRY":       "2",
}

var boolKeys = map[string]struct{}{
	"TUN_ENABLE":      {},
	"TUN_AUTO_ROUTE":  {},
	"TUN_AUTO_DETECT": {},
	"BASE64_CONVERT":  {},
}

var portKeys = map[string]struct{}{
	"EXTERNAL_PORT": {},
	"HTTP_PORT":     {},
	"SOCKS_PORT":    {},
	"MIXED_PORT":    {},
}

type configPayload struct {
	Values map[string]string `json:"values"`
	Apply  *bool             `json:"apply"`
}

type configServer struct {
	confFile string
	envFile  string
	mu       sync.Mutex
}

func serveConfigAPI(addr, confFile, envFile string) error {
	server := &configServer{
		confFile: confFile,
		envFile:  envFile,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", server.handleConfig)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("配置 API 服务已启动: %s\n", addr)
	return httpServer.ListenAndServe()
}

func (s *configServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.Method {
	case http.MethodGet:
		values, err := s.currentConfig()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"values":   values,
			"env_file": s.envFile,
		})
	case http.MethodPost:
		var payload configPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("请求格式错误: %w", err))
			return
		}

		apply := true
		if payload.Apply != nil {
			apply = *payload.Apply
		}

		if err := s.saveConfig(payload.Values, apply); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}

		values, err := s.currentConfig()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "保存成功",
			"values":  values,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *configServer) currentConfig() (map[string]string, error) {
	fileValues, err := readEnvFile(s.envFile)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string)
	for _, key := range configurableKeys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			values[key] = value
			continue
		}

		if value := strings.TrimSpace(fileValues[key]); value != "" {
			values[key] = value
			continue
		}

		if defaultValue, ok := defaultConfigValues[key]; ok {
			values[key] = defaultValue
		} else {
			values[key] = ""
		}
	}

	return values, nil
}

func (s *configServer) saveConfig(raw map[string]string, apply bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileValues, err := readEnvFile(s.envFile)
	if err != nil {
		return err
	}

	for key, value := range raw {
		if !isConfigurableKey(key) {
			return fmt.Errorf("不支持的配置项: %s", key)
		}

		normalized, err := normalizeConfigValue(key, value)
		if err != nil {
			return err
		}

		if normalized == "" {
			delete(fileValues, key)
			_ = os.Unsetenv(key)
			continue
		}

		fileValues[key] = normalized
		_ = os.Setenv(key, normalized)
	}

	if err := writeEnvFile(s.envFile, fileValues); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	if !apply {
		return nil
	}

	if strings.TrimSpace(getEnvDefault("CONF_URL", "")) == "" {
		return fmt.Errorf("CONF_URL 不能为空")
	}

	if err := updateConfig(s.confFile); err != nil {
		return fmt.Errorf("应用配置失败: %w", err)
	}

	return nil
}

func normalizeConfigValue(key, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	if _, ok := boolKeys[key]; ok {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "on":
			return "true", nil
		case "0", "false", "no", "off":
			return "false", nil
		default:
			return "", fmt.Errorf("%s 需要布尔值(true/false)", key)
		}
	}

	if _, ok := portKeys[key]; ok {
		port, err := parsePort(value)
		if err != nil {
			return "", fmt.Errorf("%s 无效: %w", key, err)
		}
		return fmt.Sprintf("%d", port), nil
	}

	if key == "CONF_TIMEOUT_SEC" || key == "CONF_RETRY" {
		num, err := strconv.Atoi(value)
		if err != nil || num < 0 {
			return "", fmt.Errorf("%s 必须是非负整数", key)
		}
		return strconv.Itoa(num), nil
	}

	return value, nil
}

func isConfigurableKey(key string) bool {
	for _, item := range configurableKeys {
		if item == key {
			return true
		}
	}
	return false
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func respondError(w http.ResponseWriter, code int, err error) {
	respondJSON(w, code, map[string]string{
		"error": err.Error(),
	})
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
