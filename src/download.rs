use crate::envfile::{get_env_bool, get_env_default, get_env_int, parse_port};
use base64::prelude::*;
use reqwest::header::USER_AGENT;
use serde_yaml::{Mapping, Value};
use std::fs;
use std::path::Path;
use std::time::Duration;

const DEFAULT_FETCH_TIMEOUT_SEC: i32 = 15;
const DEFAULT_FETCH_RETRY: i32 = 2;

pub async fn download_config(conf_file: &str) -> anyhow::Result<()> {
    let conf_url = std::env::var("CONF_URL")
        .unwrap_or_default()
        .trim()
        .to_string();
    if conf_url.is_empty() || conf_url == "http://test.com" {
        return Err(anyhow::anyhow!(
            "错误: CONF_URL 环境变量未设置或使用了默认值\n\n请在 docker-compose.yml 中配置你的订阅地址:\n  environment:\n    - CONF_URL=https://your-subscription-url"
        ));
    }

    let p = Path::new(conf_file);
    if let Some(parent) = p.parent() {
        fs::create_dir_all(parent)
            .map_err(|e| anyhow::anyhow!("创建目录失败: {}", e))?;
    }

    println!("从 {} 下载配置文件并写入到 {}", conf_url, conf_file);

    let mut data = fetch_config(&conf_url)
        .await
        .map_err(|e| anyhow::anyhow!("配置文件下载失败: {}", e))?;

    if get_env_bool("BASE64_CONVERT", false) {
        println!("配置文件转换中...");
        data = decode_base64(&data)
            .map_err(|e| anyhow::anyhow!("base64 解码失败: {}", e))?;
    }

    let mut config: Value = serde_yaml::from_slice(&data)
        .map_err(|e| anyhow::anyhow!("YAML 解析失败: {}", e))?;

    if !config.is_mapping() {
        return Err(anyhow::anyhow!("YAML 根节点必须是 Mapping 对象"));
    }

    modify_config(&mut config);

    let custom_conf = std::env::var("CUSTOM_CONF").unwrap_or_default();
    if !custom_conf.trim().is_empty() {
        if let Err(e) = merge_custom_rules(&mut config, &custom_conf) {
            println!("警告: 合并自定义规则失败: {}", e);
        }
    }

    let output = serde_yaml::to_string(&config)
        .map_err(|e| anyhow::anyhow!("YAML 序列化失败: {}", e))?;

    fs::write(conf_file, output)
        .map_err(|e| anyhow::anyhow!("写入配置文件失败: {}", e))?;

    let index_path = "/root/.config/mihomo/ui/index.html";
    if let Err(e) = crate::render::render_index(index_path) {
        println!("警告: 渲染导航页失败: {}", e);
    }

    println!("配置文件处理完成");
    Ok(())
}

async fn fetch_config(url: &str) -> anyhow::Result<Vec<u8>> {
    if let Some(file_path) = url.strip_prefix("file://") {
        return fs::read(file_path).map_err(|e| anyhow::anyhow!("{}", e));
    }

    let timeout_sec = get_env_int("CONF_TIMEOUT_SEC", DEFAULT_FETCH_TIMEOUT_SEC).max(1) as u64;
    let mut retry_count = get_env_int("CONF_RETRY", DEFAULT_FETCH_RETRY);
    if retry_count < 0 {
        retry_count = DEFAULT_FETCH_RETRY;
    }

    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(timeout_sec))
        .build()?;

    let user_agent = get_env_default("CLASH_SUB_UA", "Wget/1.21.3");

    let mut last_err = anyhow::anyhow!("未知下载错误");
    for attempt in 0..=retry_count {
        let resp_res = client
            .get(url)
            .header(USER_AGENT, &user_agent)
            .send()
            .await;

        match resp_res {
            Ok(resp) => {
                let status = resp.status();
                if status.is_success() {
                    let bytes = resp.bytes().await?;
                    return Ok(bytes.to_vec());
                } else {
                    let body_text = resp.text().await.unwrap_or_default();
                    let truncated: String = body_text.chars().take(200).collect();
                    last_err = anyhow::anyhow!("HTTP 状态码: {}, 响应: {}", status.as_u16(), truncated);
                }
            }
            Err(e) => {
                last_err = anyhow::anyhow!("{}", e);
            }
        }

        if attempt < retry_count {
            tokio::time::sleep(Duration::from_secs((attempt as u64) + 1)).await;
        }
    }

    Err(last_err)
}

fn decode_base64(data: &[u8]) -> anyhow::Result<Vec<u8>> {
    let content = String::from_utf8_lossy(data).trim().to_string();

    if let Ok(decoded) = BASE64_STANDARD.decode(&content) {
        return Ok(decoded);
    }
    if let Ok(decoded) = BASE64_STANDARD_NO_PAD.decode(&content) {
        return Ok(decoded);
    }
    if let Ok(decoded) = BASE64_URL_SAFE.decode(&content) {
        return Ok(decoded);
    }
    if let Ok(decoded) = BASE64_URL_SAFE_NO_PAD.decode(&content) {
        return Ok(decoded);
    }

    Err(anyhow::anyhow!("Base64 解码失败"))
}

fn modify_config(config: &mut Value) {
    let map = match config.as_mapping_mut() {
        Some(m) => m,
        None => return,
    };

    let external_bind = get_env_default("EXTERNAL_BIND", "0.0.0.0");
    let external_port = get_env_default("EXTERNAL_PORT", "9090");

    map.insert(
        Value::String("external-controller".into()),
        Value::String(format!("{}:{}", external_bind, external_port)),
    );
    map.insert(
        Value::String("external-ui".into()),
        Value::String("/root/.config/mihomo/ui".into()),
    );

    if let Ok(secret) = std::env::var("EXTERNAL_SECRET") {
        if !secret.trim().is_empty() {
            map.insert(Value::String("secret".into()), Value::String(secret));
        }
    }

    map.insert(Value::String("allow-lan".into()), Value::Bool(true));

    // 删除旧配置
    map.remove(&Value::String("port".into()));
    map.remove(&Value::String("socks-port".into()));
    map.remove(&Value::String("mixed-port".into()));

    let mut configured_port = false;

    if let Ok(http_port_str) = std::env::var("HTTP_PORT") {
        if !http_port_str.trim().is_empty() {
            match parse_port(&http_port_str) {
                Ok(port) => {
                    map.insert(Value::String("port".into()), Value::Number(port.into()));
                    println!("设置 HTTP 代理端口: {}", port);
                    configured_port = true;
                }
                Err(e) => eprintln!("警告: HTTP_PORT 无效: {}", e),
            }
        }
    }

    if let Ok(socks_port_str) = std::env::var("SOCKS_PORT") {
        if !socks_port_str.trim().is_empty() {
            match parse_port(&socks_port_str) {
                Ok(port) => {
                    map.insert(Value::String("socks-port".into()), Value::Number(port.into()));
                    println!("设置 SOCKS5 代理端口: {}", port);
                    configured_port = true;
                }
                Err(e) => eprintln!("警告: SOCKS_PORT 无效: {}", e),
            }
        }
    }

    if let Ok(mixed_port_str) = std::env::var("MIXED_PORT") {
        if !mixed_port_str.trim().is_empty() {
            match parse_port(&mixed_port_str) {
                Ok(port) => {
                    map.insert(Value::String("mixed-port".into()), Value::Number(port.into()));
                    println!("设置 Mixed 代理端口: {}", port);
                    configured_port = true;
                }
                Err(e) => eprintln!("警告: MIXED_PORT 无效: {}", e),
            }
        }
    }

    let has_http_env = std::env::var("HTTP_PORT").is_ok();
    let has_socks_env = std::env::var("SOCKS_PORT").is_ok();
    let has_mixed_env = std::env::var("MIXED_PORT").is_ok();
    if !has_http_env && !has_socks_env && !has_mixed_env && !configured_port {
        map.insert(Value::String("mixed-port".into()), Value::Number(7890.into()));
        println!("未显式配置代理端口，回退设置 Mixed 端口: 7890");
    }

    if get_env_bool("TUN_ENABLE", false) {
        let mut tun_map = Mapping::new();
        tun_map.insert(Value::String("enable".into()), Value::Bool(true));
        tun_map.insert(
            Value::String("stack".into()),
            Value::String(get_env_default("TUN_STACK", "system")),
        );
        tun_map.insert(
            Value::String("auto-route".into()),
            Value::Bool(get_env_bool("TUN_AUTO_ROUTE", true)),
        );
        tun_map.insert(
            Value::String("auto-detect-interface".into()),
            Value::Bool(get_env_bool("TUN_AUTO_DETECT", true)),
        );
        tun_map.insert(
            Value::String("dns-hijack".into()),
            Value::Sequence(vec![Value::String("any:53".into())]),
        );
        map.insert(Value::String("tun".into()), Value::Mapping(tun_map));
        println!("启用 TUN 模式");
    }

    println!("修改配置文件中的ui界面指向路径");
}

fn merge_custom_rules(config: &mut Value, custom_conf_path: &str) -> anyhow::Result<()> {
    let p = Path::new(custom_conf_path);
    if !p.exists() {
        return Ok(());
    }

    println!("合并自定义规则文件... {}", custom_conf_path);

    let data = fs::read(p)?;
    let custom_val: Value = serde_yaml::from_slice(&data)?;

    if let Some(custom_map) = custom_val.as_mapping() {
        if let Some(Value::Sequence(custom_rules)) = custom_map.get(&Value::String("rules".into())) {
            let config_map = config
                .as_mapping_mut()
                .ok_or_else(|| anyhow::anyhow!("config 不是 mapping"))?;

            let rules_key = Value::String("rules".into());
            let existing_rules = match config_map.get_mut(&rules_key) {
                Some(Value::Sequence(seq)) => seq,
                _ => {
                    config_map.insert(rules_key.clone(), Value::Sequence(Vec::new()));
                    match config_map.get_mut(&rules_key) {
                        Some(Value::Sequence(seq)) => seq,
                        _ => unreachable!(),
                    }
                }
            };

            for r in custom_rules {
                existing_rules.push(r.clone());
            }
        }
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_modify_config_default_ports() {
        let yaml_str = "mode: rule\nport: 1234\n";
        let mut config: Value = serde_yaml::from_str(yaml_str).unwrap();

        modify_config(&mut config);

        let map = config.as_mapping().unwrap();
        assert_eq!(
            map.get(&Value::String("external-controller".into())),
            Some(&Value::String("0.0.0.0:9090".into()))
        );
        assert_eq!(
            map.get(&Value::String("allow-lan".into())),
            Some(&Value::Bool(true))
        );
        assert_eq!(
            map.get(&Value::String("mixed-port".into())),
            Some(&Value::Number(7890.into()))
        );
        assert!(!map.contains_key(&Value::String("port".into())));
    }

    #[test]
    fn test_merge_custom_rules() {
        let tmp_dir = std::env::temp_dir();
        let custom_file = tmp_dir.join("test_custom.yaml");
        fs::write(&custom_file, "rules:\n  - MATCH,DIRECT\n").unwrap();

        let mut config: Value = serde_yaml::from_str("rules:\n  - DOMAIN,google.com,PROXY\n").unwrap();
        merge_custom_rules(&mut config, custom_file.to_str().unwrap()).unwrap();

        let map = config.as_mapping().unwrap();
        let rules = map.get(&Value::String("rules".into())).unwrap().as_sequence().unwrap();
        assert_eq!(rules.len(), 2);
        assert_eq!(rules[0], Value::String("DOMAIN,google.com,PROXY".into()));
        assert_eq!(rules[1], Value::String("MATCH,DIRECT".into()));

        let _ = fs::remove_file(custom_file);
    }

    #[test]
    fn test_decode_base64() {
        let raw = b"hello world mihomo";
        let encoded = BASE64_STANDARD.encode(raw);
        let decoded = decode_base64(encoded.as_bytes()).unwrap();
        assert_eq!(decoded, raw);
    }
}
