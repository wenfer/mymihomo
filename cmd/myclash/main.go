package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "dev"

func main() {
	// 子命令
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "download":
		downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
		confFile := downloadCmd.String("o", "/root/conf/config.yaml", "配置文件输出路径")
		downloadCmd.Parse(os.Args[2:])
		if err := downloadConfig(*confFile); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "update":
		updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
		confFile := updateCmd.String("c", "/root/conf/config.yaml", "配置文件路径")
		updateCmd.Parse(os.Args[2:])
		if err := updateConfig(*confFile); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "render":
		renderCmd := flag.NewFlagSet("render", flag.ExitOnError)
		indexFile := renderCmd.String("o", "/root/.config/mihomo/ui/index.html", "导航页输出路径")
		renderCmd.Parse(os.Args[2:])
		if err := renderIndex(*indexFile); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "version":
		fmt.Printf("myclash %s\n", version)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("用法: myclash <命令> [选项]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  download  下载并处理配置文件")
	fmt.Println("  update    更新运行中的mihomo配置")
	fmt.Println("  render    渲染导航页")
	fmt.Println("  version   显示版本信息")
}
