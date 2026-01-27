package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

type PageData struct {
	ExternalPort   string
	ExternalSecret string
	HTTPPort       string
	SocksPort      string
	MixedPort      string
	TUNEnable      bool
	TUNStack       string
	UpdateTime     string
	CronExpression string
}

const indexTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>myclash 导航</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
            padding: 40px;
            max-width: 480px;
            width: 100%;
        }
        h1 {
            color: #333;
            margin-bottom: 24px;
            font-size: 28px;
            text-align: center;
        }
        .button-container {
            display: flex;
            flex-direction: column;
            gap: 12px;
            margin-bottom: 32px;
        }
        .btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 14px 24px;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s ease;
            text-decoration: none;
            text-align: center;
        }
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
        }
        .btn-secondary {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
        }
        .btn-secondary:hover {
            box-shadow: 0 8px 20px rgba(240, 147, 251, 0.4);
        }
        .info-section {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 20px;
        }
        .info-section h2 {
            color: #555;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 16px;
        }
        .info-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
        }
        .info-item {
            background: white;
            padding: 12px;
            border-radius: 8px;
            border: 1px solid #eee;
        }
        .info-item.full-width {
            grid-column: 1 / -1;
        }
        .info-label {
            color: #888;
            font-size: 12px;
            margin-bottom: 4px;
        }
        .info-value {
            color: #333;
            font-size: 14px;
            font-weight: 500;
            font-family: 'Monaco', 'Consolas', monospace;
        }
        .status-on { color: #10b981; }
        .status-off { color: #9ca3af; }
    </style>
</head>
<body>
    <div class="container">
        <h1>myclash</h1>
        <div class="button-container">
            <a href="dashboard/#/proxies?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}" class="btn">Dashboard 面板</a>
            <a href="yacd?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}" class="btn btn-secondary">YACD 面板</a>
        </div>
        <div class="info-section">
            <h2>运行状态</h2>
            <div class="info-grid">
                {{if .HTTPPort}}
                <div class="info-item">
                    <div class="info-label">HTTP</div>
                    <div class="info-value">:{{.HTTPPort}}</div>
                </div>
                {{end}}
                {{if .SocksPort}}
                <div class="info-item">
                    <div class="info-label">SOCKS5</div>
                    <div class="info-value">:{{.SocksPort}}</div>
                </div>
                {{end}}
                {{if .MixedPort}}
                <div class="info-item">
                    <div class="info-label">Mixed</div>
                    <div class="info-value">:{{.MixedPort}}</div>
                </div>
                {{end}}
                <div class="info-item">
                    <div class="info-label">TUN 模式</div>
                    <div class="info-value {{if .TUNEnable}}status-on{{else}}status-off{{end}}">
                        {{if .TUNEnable}}启用 ({{.TUNStack}}){{else}}未启用{{end}}
                    </div>
                </div>
                <div class="info-item full-width">
                    <div class="info-label">定时更新</div>
                    <div class="info-value">{{.CronExpression}}</div>
                </div>
                <div class="info-item full-width">
                    <div class="info-label">配置更新时间</div>
                    <div class="info-value">{{.UpdateTime}}</div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`

func renderIndex(outputPath string) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	data := PageData{
		ExternalPort:   getEnvDefault("EXTERNAL_PORT", "9090"),
		ExternalSecret: os.Getenv("EXTERNAL_SECRET"),
		HTTPPort:       os.Getenv("HTTP_PORT"),
		SocksPort:      os.Getenv("SOCKS_PORT"),
		MixedPort:      os.Getenv("MIXED_PORT"),
		TUNEnable:      os.Getenv("TUN_ENABLE") == "true",
		TUNStack:       getEnvDefault("TUN_STACK", "system"),
		CronExpression: getEnvDefault("CRON_EXPRESSION", "1 * * * *"),
		UpdateTime:     time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("渲染模板失败: %w", err)
	}

	fmt.Printf("导航页已更新: %s\n", outputPath)
	return nil
}
