package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func validateCronExpression(expr string) error {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil
	}
	fields := strings.Fields(trimmed)
	if len(fields) != 5 {
		return fmt.Errorf("CRON_EXPRESSION 必须为 5 段标准格式（分 时 日 月 周），当前为 %d 段: %q", len(fields), expr)
	}
	return nil
}

func isCrondRunning() bool {
	if err := exec.Command("pidof", "crond").Run(); err == nil {
		return true
	}
	if err := exec.Command("pgrep", "-x", "crond").Run(); err == nil {
		return true
	}
	return false
}

func ensureCrondRunning() {
	if isCrondRunning() {
		return
	}

	cmd := exec.Command("crond")
	if err := cmd.Start(); err != nil {
		logWarn("启动 crond 服务失败: %v", err)
	} else {
		logInfo("crond 守护进程已启动")
	}
}

func syncCrontab() error {
	confURL := strings.TrimSpace(os.Getenv("CONF_URL"))
	cronExpr := strings.TrimSpace(os.Getenv("CRON_EXPRESSION"))
	confFile := getEnvDefault("CONF_FILE", "/root/conf/config.yaml")
	cronLog := getEnvDefault("CRON_LOG", "/root/conf/cron_history")

	// 引导模式（无有效订阅）或用户显式清空了定时表达式，则清理 crontab
	if confURL == "" || confURL == "http://test.com" || cronExpr == "" {
		_ = exec.Command("crontab", "-r").Run()
		if cronExpr == "" {
			logInfo("CRON_EXPRESSION 为空，已停用定时更新任务")
		} else {
			logInfo("当前处于引导模式（未配置有效订阅），已跳过定时更新任务")
		}
		return nil
	}

	if err := validateCronExpression(cronExpr); err != nil {
		return err
	}

	// 确保日志所在目录存在
	if err := os.MkdirAll(filepath.Dir(cronLog), 0755); err != nil {
		return fmt.Errorf("创建定时日志目录失败: %w", err)
	}

	ensureCrondRunning()

	// 写入 crontab: 分 时 日 月 周 mymihomo update -c <confFile> >> <cronLog> 2>&1
	cronEntry := fmt.Sprintf("%s mymihomo update -c %s >> %s 2>&1\n", cronExpr, confFile, cronLog)
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(cronEntry)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("更新 crontab 失败: %w, 输出: %s", err, string(out))
	}

	logInfo("定时任务已同步并生效: %s", cronExpr)
	return nil
}
