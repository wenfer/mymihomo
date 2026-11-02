package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func updateConfig(confFile string) error {
	mihomoHome := getEnvDefault("MIHOMO_HOME", "/root/.config/mihomo")
	externalPort := getEnvDefault("EXTERNAL_PORT", "9090")
	cronLog := getEnvDefault("CRON_LOG", "/root/conf/cron_history")

	// 日志文件大小控制（大于 2MB 时保留最新 2000 行）
	defer rotateLogFile(cronLog, 2*1024*1024)

	// 文件锁，防止定时任务与 Web 操作并发覆写
	lockFile := confFile + ".lock"
	if err := acquireFileLock(lockFile, 2*time.Minute); err != nil {
		return err
	}
	defer releaseFileLock(lockFile)

	logInfo("开始更新订阅配置...")

	// 1. 先下载并解析到临时文件
	tmpFile := confFile + ".download"
	defer os.Remove(tmpFile)

	if err := downloadConfig(tmpFile); err != nil {
		return fmt.Errorf("下载新配置失败: %w", err)
	}

	// 2. 语法预检：通过 /mihomo -t 确保配置能被正常加载
	if err := testMihomoConfig(mihomoHome, tmpFile); err != nil {
		return fmt.Errorf("配置预检校验未通过，已放弃更新以保护已有配置: %w", err)
	}
	logInfo("新配置文件预检校验通过: %s", tmpFile)

	// 3. 校验通过，原子替换到正式配置文件
	if err := os.Rename(tmpFile, confFile); err != nil {
		if err := copyFile(tmpFile, confFile); err != nil {
			return fmt.Errorf("写入正式配置文件失败: %w", err)
		}
	}
	logInfo("配置文件已成功更新到: %s", confFile)

	// 4. 调用 API 通知 mihomo 热重载
	url := fmt.Sprintf("http://127.0.0.1:%s/configs?force=true", externalPort)
	payload := map[string]string{"path": confFile}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建热重载请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if secret := os.Getenv("EXTERNAL_SECRET"); secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("调用热重载 API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("热重载失败, HTTP 状态码: %d, 响应: %.200s", resp.StatusCode, respBody)
	}

	logInfo("Mihomo 配置热重载成功")
	return nil
}

func testMihomoConfig(mihomoHome, testConfFile string) error {
	mihomoBin := "/mihomo"
	if _, err := os.Stat(mihomoBin); os.IsNotExist(err) {
		if p, err := exec.LookPath("mihomo"); err == nil {
			mihomoBin = p
		} else {
			// 当前环境无 mihomo 可执行文件时（如非容器本地测试），跳过校验
			return nil
		}
	}

	cmd := exec.Command(mihomoBin, "-d", mihomoHome, "-f", testConfFile, "-t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("校验失败: %w, 详情: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func acquireFileLock(lockFile string, maxAge time.Duration) error {
	if fi, err := os.Stat(lockFile); err == nil {
		if time.Since(fi.ModTime()) > maxAge {
			_ = os.Remove(lockFile)
		} else {
			return fmt.Errorf("已有更新任务正在运行中 (lock: %s)", lockFile)
		}
	}

	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("获取更新锁失败: %w", err)
	}
	_ = f.Close()
	return nil
}

func releaseFileLock(lockFile string) {
	_ = os.Remove(lockFile)
}

func rotateLogFile(filePath string, maxBytes int64) {
	fi, err := os.Stat(filePath)
	if err != nil || fi.Size() < maxBytes {
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 2000 {
		lines = lines[len(lines)-2000:]
		_ = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	}
}
