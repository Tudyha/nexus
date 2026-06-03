package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nexus-cli",
	Short: "Nexus 内网穿透客户端",
	Long: `Nexus 内网穿透客户端 — 将内网服务通过公网服务器暴露到外网。

使用前需要先在服务端管理后台创建应用（App），获取 app_id 和 app_secret。
然后通过 --config-file (-f) 或 --config (-c) 指定配置信息启动客户端。

示例:
  # 使用配置文件启动（前台）
  nexus-cli run -f config.json

  # 使用 base64 配置启动（前台）
  nexus-cli run -c $(base64 config.json)

  # 后台守护进程模式
  nexus-cli run -d -f config.json

  # 查看运行状态
  nexus-cli status

  # 查看日志
  nexus-cli logs

  # 停止客户端
  nexus-cli stop
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
