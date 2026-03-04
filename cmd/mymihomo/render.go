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
	WebConfigPort  string
}

const indexTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>mymihomo 控制中心</title>
    <style>
        :root {
            --bg-1: #f2efe8;
            --bg-2: #dbe8e6;
            --card: #ffffff;
            --ink: #14212b;
            --sub: #52606d;
            --line: #d5dde5;
            --brand: #0f766e;
            --brand-2: #0b4f6c;
            --warn: #b45309;
            --ok: #166534;
            --bad: #b91c1c;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: "IBM Plex Sans", "Noto Sans SC", "PingFang SC", sans-serif;
            background: radial-gradient(circle at 20% 20%, var(--bg-2), var(--bg-1) 52%);
            color: var(--ink);
            min-height: 100vh;
            padding: 24px 16px 56px;
        }
        .wrap {
            max-width: 980px;
            margin: 0 auto;
        }
        .hero {
            margin-bottom: 18px;
            padding: 22px;
            border-radius: 16px;
            background: linear-gradient(120deg, rgba(15,118,110,0.16), rgba(11,79,108,0.14));
            border: 1px solid rgba(15,118,110,0.2);
        }
        .hero h1 { font-size: 30px; margin-bottom: 8px; letter-spacing: 0.5px; }
        .hero p { color: var(--sub); font-size: 14px; }
        .grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 14px;
        }
        .card {
            background: var(--card);
            border: 1px solid var(--line);
            border-radius: 16px;
            padding: 18px;
            box-shadow: 0 10px 24px rgba(20, 33, 43, 0.06);
        }
        .card h2 {
            font-size: 17px;
            margin-bottom: 12px;
            letter-spacing: 0.3px;
        }
        .muted {
            color: var(--sub);
            font-size: 13px;
            line-height: 1.55;
            margin-bottom: 12px;
        }
        .nav-row {
            display: grid;
            grid-template-columns: repeat(3, minmax(0, 1fr));
            gap: 10px;
            margin-bottom: 12px;
        }
        .btn {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 11px 14px;
            border-radius: 10px;
            border: 1px solid transparent;
            text-decoration: none;
            font-weight: 600;
            font-size: 14px;
            cursor: pointer;
            transition: transform .12s ease, box-shadow .2s ease, border-color .2s ease;
        }
        .btn:hover { transform: translateY(-1px); }
        .btn-main {
            background: linear-gradient(120deg, var(--brand), var(--brand-2));
            color: #fff;
            box-shadow: 0 8px 18px rgba(15, 118, 110, 0.28);
        }
        .btn-soft {
            background: #f3f9f8;
            color: #0f4d47;
            border-color: #b7e0da;
        }
        .btn-alt {
            background: #fef7ed;
            color: #9a3412;
            border-color: #fed7aa;
        }
        .status {
            border-radius: 10px;
            padding: 10px 12px;
            font-size: 13px;
            line-height: 1.45;
            border: 1px solid;
            margin-bottom: 12px;
            background: #f8fafc;
            color: #334155;
            border-color: #cbd5e1;
        }
        .status.ok { color: var(--ok); border-color: #bbf7d0; background: #f0fdf4; }
        .status.err { color: var(--bad); border-color: #fecaca; background: #fef2f2; }
        .status.warn { color: var(--warn); border-color: #fde68a; background: #fffbeb; }
        form {
            display: grid;
            grid-template-columns: repeat(2, minmax(0, 1fr));
            gap: 10px;
        }
        .field { display: flex; flex-direction: column; gap: 5px; }
        .field.full { grid-column: 1 / -1; }
        label { font-size: 12px; color: #334155; font-weight: 600; letter-spacing: 0.2px; }
        input, select {
            width: 100%;
            border: 1px solid #cbd5e1;
            border-radius: 8px;
            padding: 9px 10px;
            font-size: 14px;
            font-family: "IBM Plex Sans", "Noto Sans SC", sans-serif;
            color: var(--ink);
            background: #fff;
        }
        input:focus, select:focus {
            outline: none;
            border-color: #0f766e;
            box-shadow: 0 0 0 3px rgba(15, 118, 110, 0.12);
        }
        .check { display: flex; align-items: center; gap: 8px; margin-top: 8px; }
        .check input { width: auto; }
        .actions {
            grid-column: 1 / -1;
            display: flex;
            gap: 10px;
            padding-top: 6px;
        }
        .meta {
            margin-top: 12px;
            font-size: 12px;
            color: #64748b;
            line-height: 1.6;
        }
        @media (max-width: 860px) {
            .nav-row, form { grid-template-columns: 1fr; }
            .hero h1 { font-size: 26px; }
        }
    </style>
</head>
<body>
    <div class="wrap">
        <section class="hero">
            <h1>mymihomo 控制中心</h1>
            <p>导航 + 在线配置。你可以直接在这里维护订阅和端口配置，再一键应用到运行中的 mihomo。</p>
        </section>

        <div class="grid">
            <section class="card">
                <h2>面板导航</h2>
                <p class="muted">保留原有导航能力，链接会根据当前配置自动更新。</p>
                <div class="nav-row">
                    <a id="dashboard-link" class="btn btn-main" href="dashboard/#/proxies?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}">Dashboard</a>
                    <a id="yacd-link" class="btn btn-soft" href="yacd?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}">YACD</a>
                    <a id="zashboard-link" class="btn btn-alt" href="zashboard/?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}">zashboard</a>
                </div>
                <div class="meta">
                    当前端口: {{.ExternalPort}} | Mixed: {{.MixedPort}} | 更新于: {{.UpdateTime}}
                </div>
            </section>

            <section class="card">
                <h2>运行配置</h2>
                <div id="status" class="status">正在加载配置...</div>
                <form id="config-form">
                    <div class="field full">
                        <label for="CONF_URL">订阅地址 CONF_URL</label>
                        <input id="CONF_URL" name="CONF_URL" type="text" placeholder="https://your-subscription-url" required />
                    </div>

                    <div class="field">
                        <label for="EXTERNAL_BIND">API 绑定地址</label>
                        <input id="EXTERNAL_BIND" name="EXTERNAL_BIND" type="text" />
                    </div>
                    <div class="field">
                        <label for="EXTERNAL_PORT">API 端口</label>
                        <input id="EXTERNAL_PORT" name="EXTERNAL_PORT" type="number" min="1" max="65535" />
                    </div>

                    <div class="field full">
                        <label for="EXTERNAL_SECRET">API 密钥</label>
                        <input id="EXTERNAL_SECRET" name="EXTERNAL_SECRET" type="text" />
                    </div>

                    <div class="field">
                        <label for="MIXED_PORT">Mixed 端口</label>
                        <input id="MIXED_PORT" name="MIXED_PORT" type="number" min="1" max="65535" />
                    </div>
                    <div class="field">
                        <label for="HTTP_PORT">HTTP 端口</label>
                        <input id="HTTP_PORT" name="HTTP_PORT" type="number" min="1" max="65535" />
                    </div>

                    <div class="field">
                        <label for="SOCKS_PORT">SOCKS5 端口</label>
                        <input id="SOCKS_PORT" name="SOCKS_PORT" type="number" min="1" max="65535" />
                    </div>
                    <div class="field">
                        <label for="TUN_STACK">TUN 协议栈</label>
                        <select id="TUN_STACK" name="TUN_STACK">
                            <option value="system">system</option>
                            <option value="gvisor">gvisor</option>
                            <option value="mixed">mixed</option>
                        </select>
                    </div>

                    <div class="field">
                        <label for="CRON_EXPRESSION">定时更新表达式</label>
                        <input id="CRON_EXPRESSION" name="CRON_EXPRESSION" type="text" placeholder="1 * * * *" />
                    </div>
                    <div class="field">
                        <label for="CONF_TIMEOUT_SEC">下载超时(秒)</label>
                        <input id="CONF_TIMEOUT_SEC" name="CONF_TIMEOUT_SEC" type="number" min="1" />
                    </div>

                    <div class="field">
                        <label for="CONF_RETRY">下载重试次数</label>
                        <input id="CONF_RETRY" name="CONF_RETRY" type="number" min="0" />
                    </div>
                    <div class="field">
                        <label for="CUSTOM_CONF">自定义规则文件</label>
                        <input id="CUSTOM_CONF" name="CUSTOM_CONF" type="text" placeholder="/root/conf/custom.yaml" />
                    </div>

                    <div class="field full">
                        <label class="check"><input id="TUN_ENABLE" name="TUN_ENABLE" type="checkbox" /> 启用 TUN 模式</label>
                        <label class="check"><input id="BASE64_CONVERT" name="BASE64_CONVERT" type="checkbox" /> 订阅启用 Base64 解码</label>
                    </div>

                    <div class="actions">
                        <button type="submit" class="btn btn-main">保存并应用</button>
                        <button id="save-only" type="button" class="btn btn-soft">仅保存</button>
                        <button id="reload" type="button" class="btn btn-soft">重新加载</button>
                    </div>
                </form>
                <div class="meta">配置接口: <code id="api-base"></code>。如果加载失败，请在容器端口映射中暴露 WEB_CONFIG_PORT（默认 18080）。</div>
            </section>
        </div>
    </div>

    <script>
        const API_BASE = location.protocol + '//' + location.hostname + ':{{.WebConfigPort}}';
        const BOOL_FIELDS = ["TUN_ENABLE", "BASE64_CONVERT"];
        const TEXT_FIELDS = [
            "CONF_URL", "EXTERNAL_BIND", "EXTERNAL_PORT", "EXTERNAL_SECRET", "HTTP_PORT", "SOCKS_PORT",
            "MIXED_PORT", "TUN_STACK", "CRON_EXPRESSION", "CONF_TIMEOUT_SEC", "CONF_RETRY", "CUSTOM_CONF"
        ];

        document.getElementById("api-base").textContent = API_BASE;

        function setStatus(msg, type = "") {
            const el = document.getElementById("status");
            el.className = ('status ' + type).trim();
            el.textContent = msg;
        }

        function fillForm(values = {}) {
            for (const key of TEXT_FIELDS) {
                const el = document.getElementById(key);
                if (el) el.value = values[key] || "";
            }
            for (const key of BOOL_FIELDS) {
                const el = document.getElementById(key);
                if (el) el.checked = (values[key] || "").toString() === "true";
            }
            updateNavLinks(values);
        }

        function readForm() {
            const values = {};
            for (const key of TEXT_FIELDS) {
                const el = document.getElementById(key);
                values[key] = el ? el.value.trim() : "";
            }
            for (const key of BOOL_FIELDS) {
                const el = document.getElementById(key);
                values[key] = el && el.checked ? "true" : "false";
            }
            return values;
        }

        function updateNavLinks(values) {
            const port = values.EXTERNAL_PORT || "9090";
            const secret = encodeURIComponent(values.EXTERNAL_SECRET || "");
            document.getElementById("dashboard-link").href = 'dashboard/#/proxies?hostname=' + port + '&secret=' + secret;
            document.getElementById("yacd-link").href = 'yacd?hostname=' + port + '&secret=' + secret;
            document.getElementById("zashboard-link").href = 'zashboard/?hostname=' + port + '&secret=' + secret;
        }

        async function loadConfig() {
            setStatus("正在读取配置...", "");
            try {
                const res = await fetch(API_BASE + '/api/config');
                const data = await res.json();
                if (!res.ok) {
                    throw new Error(data.error || "读取失败");
                }
                fillForm(data.values || {});
                setStatus("配置读取成功", "ok");
            } catch (err) {
                setStatus('配置接口不可用: ' + err.message, "warn");
            }
        }

        async function saveConfig(apply) {
            const values = readForm();
            if (!values.CONF_URL) {
                setStatus("CONF_URL 不能为空", "err");
                return;
            }

            const btnText = apply ? "正在保存并应用..." : "正在保存...";
            setStatus(btnText, "");

            try {
                const res = await fetch(API_BASE + '/api/config', {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ values, apply }),
                });
                const data = await res.json();
                if (!res.ok) {
                    throw new Error(data.error || "保存失败");
                }
                fillForm(data.values || values);
                setStatus(apply ? "保存并应用成功" : "保存成功", "ok");
            } catch (err) {
                setStatus('操作失败: ' + err.message, "err");
            }
        }

        document.getElementById("config-form").addEventListener("submit", (e) => {
            e.preventDefault();
            saveConfig(true);
        });

        document.getElementById("save-only").addEventListener("click", () => {
            saveConfig(false);
        });

        document.getElementById("reload").addEventListener("click", () => {
            loadConfig();
        });

        loadConfig();
    </script>
</body>
</html>`

func renderIndex(outputPath string) error {
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
		WebConfigPort:  getEnvDefault("WEB_CONFIG_PORT", "18080"),
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
