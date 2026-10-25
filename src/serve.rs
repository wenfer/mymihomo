use crate::envfile::{parse_port, read_env_file, write_env_file};
use crate::update::update_config;
use axum::{
    extract::State,
    http::{header, Method, StatusCode},
    response::{IntoResponse, Response},
    routing::get,
    Json, Router,
};
use serde::Deserialize;
use serde_json::json;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::Mutex;
use tower_http::cors::{Any, CorsLayer};

pub const CONFIGURABLE_KEYS: &[&str] = &[
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
];

const BOOL_KEYS: &[&str] = &[
    "TUN_ENABLE",
    "TUN_AUTO_ROUTE",
    "TUN_AUTO_DETECT",
    "BASE64_CONVERT",
];

const PORT_KEYS: &[&str] = &["EXTERNAL_PORT", "HTTP_PORT", "SOCKS_PORT", "MIXED_PORT"];

fn default_config_values() -> HashMap<&'static str, &'static str> {
    let mut map = HashMap::new();
    map.insert("EXTERNAL_BIND", "0.0.0.0");
    map.insert("EXTERNAL_PORT", "9090");
    map.insert("MIXED_PORT", "7890");
    map.insert("TUN_ENABLE", "false");
    map.insert("TUN_STACK", "system");
    map.insert("TUN_AUTO_ROUTE", "true");
    map.insert("TUN_AUTO_DETECT", "true");
    map.insert("BASE64_CONVERT", "false");
    map.insert("CRON_EXPRESSION", "1 * * * *");
    map.insert("CONF_TIMEOUT_SEC", "15");
    map.insert("CONF_RETRY", "2");
    map
}

#[derive(Debug, Deserialize)]
pub struct ConfigPayload {
    pub values: HashMap<String, String>,
    pub apply: Option<bool>,
}

#[derive(Clone)]
pub struct AppState {
    pub conf_file: String,
    pub env_file: String,
    pub lock: Arc<Mutex<()>>,
}

pub async fn serve_config_api(addr: &str, conf_file: &str, env_file: &str) -> anyhow::Result<()> {
    let state = AppState {
        conf_file: conf_file.to_string(),
        env_file: env_file.to_string(),
        lock: Arc::new(Mutex::new(())),
    };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods([Method::GET, Method::POST, Method::OPTIONS])
        .allow_headers([header::CONTENT_TYPE, header::AUTHORIZATION]);

    let app = Router::new()
        .route("/healthz", get(healthz))
        .route("/api/config", get(get_config).post(post_config))
        .layer(cors)
        .with_state(state);

    let clean_addr = if addr.starts_with(':') {
        format!("0.0.0.0{}", addr)
    } else {
        addr.to_string()
    };

    let listener = tokio::net::TcpListener::bind(&clean_addr).await?;
    println!("配置 API 服务已启动: {}", clean_addr);

    axum::serve(listener, app).await?;
    Ok(())
}

async fn healthz() -> impl IntoResponse {
    Json(json!({ "status": "ok" }))
}

async fn get_config(State(state): State<AppState>) -> Response {
    match current_config(&state.env_file).await {
        Ok(values) => (
            StatusCode::OK,
            Json(json!({
                "values": values,
                "env_file": state.env_file,
            })),
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn post_config(State(state): State<AppState>, Json(payload): Json<ConfigPayload>) -> Response {
    let apply = payload.apply.unwrap_or(true);
    let _guard = state.lock.lock().await;

    if let Err(e) = save_config(&state.conf_file, &state.env_file, payload.values, apply).await {
        return (
            StatusCode::BAD_REQUEST,
            Json(json!({ "error": e.to_string() })),
        )
            .into_response();
    }

    match current_config(&state.env_file).await {
        Ok(values) => (
            StatusCode::OK,
            Json(json!({
                "message": "保存成功",
                "values": values,
            })),
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({ "error": e.to_string() })),
        )
            .into_response(),
    }
}

async fn current_config(env_file: &str) -> anyhow::Result<HashMap<String, String>> {
    let file_values = read_env_file(env_file)?;
    let defaults = default_config_values();
    let mut values = HashMap::new();

    for &key in CONFIGURABLE_KEYS {
        if let Ok(v) = std::env::var(key) {
            let trimmed = v.trim();
            if !trimmed.is_empty() {
                values.insert(key.to_string(), trimmed.to_string());
                continue;
            }
        }

        if let Some(v) = file_values.get(key) {
            let trimmed = v.trim();
            if !trimmed.is_empty() {
                values.insert(key.to_string(), trimmed.to_string());
                continue;
            }
        }

        if let Some(&def) = defaults.get(key) {
            values.insert(key.to_string(), def.to_string());
        } else {
            values.insert(key.to_string(), String::new());
        }
    }

    Ok(values)
}

async fn save_config(
    conf_file: &str,
    env_file: &str,
    raw: HashMap<String, String>,
    apply: bool,
) -> anyhow::Result<()> {
    let mut file_values = read_env_file(env_file)?;

    for (key, value) in raw {
        if !CONFIGURABLE_KEYS.contains(&key.as_str()) {
            return Err(anyhow::anyhow!("不支持的配置项: {}", key));
        }

        let normalized = normalize_config_value(&key, &value)?;
        if normalized.is_empty() {
            file_values.remove(&key);
            std::env::remove_var(&key);
        } else {
            file_values.insert(key.clone(), normalized.clone());
            std::env::set_var(&key, &normalized);
        }
    }

    write_env_file(env_file, &file_values)
        .map_err(|e| anyhow::anyhow!("写入配置文件失败: {}", e))?;

    if !apply {
        return Ok(());
    }

    let conf_url = std::env::var("CONF_URL").unwrap_or_default();
    if conf_url.trim().is_empty() {
        return Err(anyhow::anyhow!("CONF_URL 不能为空"));
    }

    update_config(conf_file)
        .await
        .map_err(|e| anyhow::anyhow!("应用配置失败: {}", e))?;

    Ok(())
}

fn normalize_config_value(key: &str, value: &str) -> anyhow::Result<String> {
    let trimmed = value.trim();
    if trimmed.is_empty() {
        return Ok(String::new());
    }

    if BOOL_KEYS.contains(&key) {
        match trimmed.to_lowercase().as_str() {
            "1" | "true" | "yes" | "on" => return Ok("true".to_string()),
            "0" | "false" | "no" | "off" => return Ok("false".to_string()),
            _ => return Err(anyhow::anyhow!("{} 需要布尔值(true/false)", key)),
        }
    }

    if PORT_KEYS.contains(&key) {
        let port = parse_port(trimmed).map_err(|e| anyhow::anyhow!("{} 无效: {}", key, e))?;
        return Ok(port.to_string());
    }

    if key == "CONF_TIMEOUT_SEC" || key == "CONF_RETRY" {
        match trimmed.parse::<i32>() {
            Ok(n) if n >= 0 => return Ok(n.to_string()),
            _ => return Err(anyhow::anyhow!("{} 必须是非负整数", key)),
        }
    }

    Ok(trimmed.to_string())
}
