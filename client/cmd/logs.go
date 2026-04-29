/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
)

var follow bool

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if follow {
			// 类似 tail -f 的实现
			t, err := tail.TailFile(logFile, tail.Config{
				Follow:    true,
				ReOpen:    true, // 如果文件被切割，自动重新打开
				MustExist: true,
			})
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("--- 正在实时追踪日志: %s ---\n", logFile)
			for line := range t.Lines {
				fmt.Println(line.Text)
			}
		} else {
			// 仅仅读取并打印当前内容
			content, err := os.ReadFile(logFile)
			if err != nil {
				log.Fatalf("无法读取日志文件: %v", err)
			}
			fmt.Print(string(content))
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "实时追踪日志输出")
}
