mod download;
mod envfile;
mod render;
mod serve;
mod update;

use clap::{Parser, Subcommand};
use envfile::{get_env_default, get_web_env_file, load_persisted_env};
use std::process;

const VERSION: &str = match option_env!("VERSION") {
    Some(v) => v,
    None => env!("CARGO_PKG_VERSION"),
};

#[derive(Parser)]
#[command(
    name = "mymihomo",
    version = VERSION,
    about = "基于 mihomo 的 Docker 镜像辅助工具 (Rust 版本)"
)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// 下载并处理配置文件
    Download {
        /// 配置文件输出路径
        #[arg(short = 'o', default_value = "/root/conf/config.yaml")]
        output: String,
    },
    /// 更新运行中的mihomo配置
    Update {
        /// 配置文件路径
        #[arg(short = 'c', default_value = "/root/conf/config.yaml")]
        config: String,
    },
    /// 渲染导航页
    Render {
        /// 导航页输出路径
        #[arg(short = 'o', default_value = "/root/.config/mihomo/ui/index.html")]
        output: String,
    },
    /// 显示版本信息
    Version,
    /// 启动配置 API 服务
    Serve {
        /// 配置服务监听地址
        #[arg(long, default_value_t = format!(":{}", get_env_default("WEB_CONFIG_PORT", "18080")))]
        addr: String,

        /// 配置文件路径
        #[arg(short = 'c', default_value_t = get_env_default("CONF_FILE", "/root/conf/config.yaml"))]
        config: String,

        /// 持久化环境变量文件
        #[arg(short = 'e', default_value_t = get_web_env_file())]
        env_file: String,
    },
}

#[tokio::main]
async fn main() {
    if let Err(err) = load_persisted_env() {
        eprintln!("警告: 加载持久化配置失败: {}", err);
    }

    let raw_args: Vec<String> = std::env::args()
        .enumerate()
        .map(|(i, arg)| {
            if i > 0 && arg.starts_with('-') && !arg.starts_with("--") && arg.len() > 2 {
                // 将形如 -addr 的 Go 风格单横线长参数转换为 --addr
                format!("-{}", arg)
            } else {
                arg
            }
        })
        .collect();

    let cli = Cli::parse_from(raw_args);

    match cli.command {
        Commands::Download { output } => {
            if let Err(err) = download::download_config(&output).await {
                eprintln!("错误: {}", err);
                process::exit(1);
            }
        }
        Commands::Update { config } => {
            if let Err(err) = update::update_config(&config).await {
                eprintln!("错误: {}", err);
                process::exit(1);
            }
        }
        Commands::Render { output } => {
            if let Err(err) = render::render_index(&output) {
                eprintln!("错误: {}", err);
                process::exit(1);
            }
        }
        Commands::Version => {
            println!("mymihomo {}", VERSION);
        }
        Commands::Serve {
            addr,
            config,
            env_file,
        } => {
            if let Err(err) = serve::serve_config_api(&addr, &config, &env_file).await {
                eprintln!("错误: {}", err);
                process::exit(1);
            }
        }
    }
}
