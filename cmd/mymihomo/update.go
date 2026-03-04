package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func updateConfig(confFile string) error {
	externalPort := getEnvDefault("EXTERNAL_PORT", "9090")

	// 先下载配置
	if err := downloadConfig(confFile); err != nil {
		return err
	}

	// 调用 API 更新配置
	url := fmt.Sprintf("http://127.0.0.1:%s/configs?force=true", externalPort)
	payload := map[string]string{"path": confFile}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加鉴权
	if secret := os.Getenv("EXTERNAL_SECRET"); secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("更新配置失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("更新配置失败, HTTP 状态码: %d, 响应: %.200s", resp.StatusCode, respBody)
	}

	fmt.Println("配置更新成功")
	return nil
}
