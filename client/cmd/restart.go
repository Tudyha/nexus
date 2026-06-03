package cmd

import (
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "重启客户端（停止后重新启动守护进程）",
	Long: `停止正在运行的客户端并重新以守护进程模式启动。

等效于依次执行 nexus-cli stop 和 nexus-cli run -d。

示例:
  nexus-cli restart -f config.json

注意: 重启需要传入与首次启动时相同的配置参数。
`,
	Run: func(cmd *cobra.Command, args []string) {
		stop()
		runAsDaemon()
	},
}

func init() {
	restartCmd.Flags().StringVarP(&configBase64, "config", "c", "", "Base64 encoded config JSON")
	restartCmd.Flags().StringVarP(&configFile, "config-file", "f", "", "Path to config JSON file")
	rootCmd.AddCommand(restartCmd)
}
