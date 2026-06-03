package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var follow bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "查看客户端日志",
	Long: `查看 Nexus 客户端的运行日志。

默认输出全部日志内容。
使用 --follow (-f) 选项可以实时跟踪日志输出（类似 tail -f）。

示例:
  nexus-cli logs
  nexus-cli logs -f
`,
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
			return
		}
		defer f.Close()

		if follow {
			f.Seek(0, io.SeekEnd)
			r := bufio.NewReader(f)
			for {
				line, err := r.ReadString('\n')
				fmt.Print(line)
				if err != nil {
					time.Sleep(1 * time.Second)
				}
			}
		} else {
			content, err := io.ReadAll(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read log file: %v\n", err)
				return
			}
			fmt.Print(string(content))
		}
	},
}


func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
}
