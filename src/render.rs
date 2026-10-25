use crate::envfile::get_env_default;
use chrono::Local;
use std::fs::{self, File};
use std::io::Write;
use std::path::Path;

const INDEX_TEMPLATE: &str = r#"<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>mymihomo 控制台</title>
    <style>
        :root {
            --bg: #f8fafc;
            --card: #ffffff;
            --ink: #0f172a;
            --sub: #475569;
            --line: #e2e8f0;
            --brand: #0f766e;
            --brand-hover: #115e59;
            --radius: 6px;
            --ok: #15803d;
            --bad: #b91c1c;
            --warn: #b45309;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: var(--bg);
            color: var(--ink);
            min-height: 100vh;
            padding: 32px 16px 64px;
            -webkit-font-smoothing: antialiased;
        }
        .wrap {
            max-width: 920px;
            margin: 0 auto;
        }
        .header {
            margin-bottom: 20px;
            padding-bottom: 16px;
            border-bottom: 1px solid var(--line);
        }
        .header h1 {
            font-size: 22px;
            font-weight: 600;
            letter-spacing: -0.2px;
            margin-bottom: 4px;
        }
        .header p {
            color: var(--sub);
            font-size: 13px;
        }
        .grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 16px;
        }
        .card {
            background: var(--card);
            border: 1px solid var(--line);
            border-radius: var(--radius);
            padding: 20px;
            box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
        }
        .card h2 {
            font-size: 15px;
            font-weight: 600;
            margin-bottom: 8px;
            color: var(--ink);
        }
        .muted {
            color: var(--sub);
            font-size: 13px;
            line-height: 1.5;
            margin-bottom: 14px;
        }
        .nav-row {
            display: grid;
            grid-template-columns: repeat(3, minmax(0, 1fr));
            gap: 10px;
            margin-bottom: 14px;
        }
        .btn {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 9px 14px;
            border-radius: var(--radius);
            border: 1px solid var(--line);
            text-decoration: none;
            font-weight: 500;
            font-size: 13px;
            cursor: pointer;
            transition: background-color .15s ease, border-color .15s ease;
        }
        .btn-main {
            background: var(--brand);
            border-color: var(--brand);
            color: #fff;
            box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
        }
        .btn-main:hover {
            background: var(--brand-hover);
            border-color: var(--brand-hover);
        }
        .btn-secondary {
            background: #fff;
            color: var(--ink);
            border-color: var(--line);
        }
        .btn-secondary:hover {
            background: #f1f5f9;
            border-color: #cbd5e1;
        }
        .status {
            border-radius: var(--radius);
            padding: 8px 12px;
            font-size: 13px;
            line-height: 1.4;
            border: 1px solid var(--line);
            margin-bottom: 14px;
            background: #f8fafc;
            color: var(--sub);
        }
        .status.ok {
            border-color: #bbf7d0;
            background: #f0fdf4;
            color: var(--ok);
        }
        .status.err {
            border-color: #fecaca;
            background: #fef2f2;
            color: var(--bad);
        }
        .status.warn {
            border-color: #fde68a;
            background: #fffbeb;
            color: var(--warn);
        }
        form {
            display: grid;
            grid-template-columns: repeat(2, minmax(0, 1fr));
            gap: 12px;
        }
        .field { display: flex; flex-direction: column; gap: 5px; }
        .field.full { grid-column: 1 / -1; }
        label { font-size: 12px; color: var(--sub); font-weight: 500; }
        input, select {
            width: 100%;
            border: 1px solid var(--line);
            border-radius: var(--radius);
            padding: 8px 10px;
            font-size: 13px;
            font-family: inherit;
            color: var(--ink);
            background: #fff;
            transition: border-color .15s ease, box-shadow .15s ease;
        }
        input:focus, select:focus {
            outline: none;
            border-color: var(--brand);
            box-shadow: 0 0 0 2px rgba(15, 118, 110, 0.15);
        }
        .check { display: flex; align-items: center; gap: 8px; margin-top: 6px; cursor: pointer; }
        .check input { width: auto; cursor: pointer; }
        .actions {
            grid-column: 1 / -1;
            display: flex;
            gap: 10px;
            padding-top: 8px;
        }
        .meta {
            margin-top: 12px;
            font-size: 12px;
            color: var(--sub);
            line-height: 1.5;
        }
        code {
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace; /* deslop-ignore 34 */
            background: #f1f5f9;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 12px;
            color: var(--ink);
        }
        @media (max-width: 860px) {
            .nav-row, form { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="wrap">
        <header class="header">
            <h1>mymihomo 控制台</h1>
            <p>管理节点订阅与端口映射，保存后热重载生效。</p>
        </header>

        <div class="grid">
            <section class="card">
                <h2>仪表盘入口</h2>
                <p class="muted">外部面板直达链接，自动携带当前端口与密钥：</p>
                <div class="nav-row">
                    <a id="dashboard-link" class="btn btn-secondary" href="dashboard/#/proxies?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}">Dashboard</a>
                    <a id="yacd-link" class="btn btn-secondary" href="yacd?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}">YACD</a>
                    <a id="zashboard-link" class="btn btn-secondary" href="zashboard/?hostname={{.ExternalPort}}&secret={{.ExternalSecret}}">zashboard</a>
                </div>
                <div class="meta">
                    当前端口: <code>{{.ExternalPort}}</code> | Mixed 代理: <code>{{.MixedPort}}</code> | 上次更新: {{.UpdateTime}}
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
                        <button id="save-only" type="button" class="btn btn-secondary">仅保存</button>
                        <button id="reload" type="button" class="btn btn-secondary">重新拉取</button>
                    </div>
                </form>
                <div class="meta">接口地址: <code id="api-base"></code>。如无法连通，请确认容器已暴露 <code>WEB_CONFIG_PORT</code> 映射端口。</div>
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
</html>"#;

pub fn render_index(output_path: &str) -> anyhow::Result<()> {
    let p = Path::new(output_path);
    if let Some(parent) = p.parent() {
        fs::create_dir_all(parent)?;
    }

    let external_port = get_env_default("EXTERNAL_PORT", "9090");
    let external_secret = std::env::var("EXTERNAL_SECRET").unwrap_or_default();
    let mixed_port = std::env::var("MIXED_PORT").unwrap_or_default();
    let web_config_port = get_env_default("WEB_CONFIG_PORT", "18080");
    let update_time = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();

    let rendered = INDEX_TEMPLATE
        .replace("{{.ExternalPort}}", &external_port)
        .replace("{{.ExternalSecret}}", &external_secret)
        .replace("{{.MixedPort}}", &mixed_port)
        .replace("{{.UpdateTime}}", &update_time)
        .replace("{{.WebConfigPort}}", &web_config_port);

    let mut file = File::create(p)?;
    file.write_all(rendered.as_bytes())?;

    println!("导航页已更新: {}", output_path);
    Ok(())
}
