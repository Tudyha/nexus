package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看客户端运行状态",
	Long: `查看 Nexus 客户端是否正在运行。

如果正在运行，将显示进程 PID。
如果未运行，将提示 "not running"。

示例:
  nexus-cli status
`,
	Run: func(cmd *cobra.Command, args []string) {
		if p, running := isRunning(); running {
			fmt.Printf("%s is already running with PID: %d\n", programName, p.Pid)
		} else {
			fmt.Printf("%s is not running\n", programName)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
