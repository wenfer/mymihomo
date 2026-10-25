use crate::download::download_config;
use crate::envfile::get_env_default;
use reqwest::header::{AUTHORIZATION, CONTENT_TYPE};
use serde_json::json;
use std::time::Duration;

pub async fn update_config(conf_file: &str) -> anyhow::Result<()> {
    let external_port = get_env_default("EXTERNAL_PORT", "9090");

    // 先下载配置
    download_config(conf_file).await?;

    // 调用 API 更新配置
    let url = format!("http://127.0.0.1:{}/configs?force=true", external_port);
    let payload = json!({ "path": conf_file });

    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(10))
        .build()?;

    let mut req = client
        .put(&url)
        .header(CONTENT_TYPE, "application/json")
        .json(&payload);

    if let Ok(secret) = std::env::var("EXTERNAL_SECRET") {
        if !secret.trim().is_empty() {
            req = req.header(AUTHORIZATION, format!("Bearer {}", secret.trim()));
        }
    }

    let resp = req
        .send()
        .await
        .map_err(|e| anyhow::anyhow!("更新配置失败: {}", e))?;

    let status = resp.status();
    if status != reqwest::StatusCode::OK && status != reqwest::StatusCode::NO_CONTENT {
        let resp_body = resp.text().await.unwrap_or_default();
        let truncated: String = resp_body.chars().take(200).collect();
        return Err(anyhow::anyhow!(
            "更新配置失败, HTTP 状态码: {}, 响应: {}",
            status.as_u16(),
            truncated
        ));
    }

    println!("配置更新成功");
    Ok(())
}
