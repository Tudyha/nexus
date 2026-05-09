package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tudyha/nexus/client/app"
	"github.com/Tudyha/nexus/client/config"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

var (
	programName = "nexus-cli"
	pidFile     = ".pid"
	logFile     = ".log"
	cntxt       = &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0644,
		LogFileName: logFile,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}

	daemonFlag   bool
	configBase64 string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run",
	Run: func(cmd *cobra.Command, args []string) {
		if daemonFlag {
			runAsDaemon()
		} else {
			run()
		}
	},
}

func init() {
	runCmd.Flags().BoolVarP(&daemonFlag, "daemon", "d", false, "Run in background (daemon mode)")
	runCmd.Flags().StringVarP(&configBase64, "config", "c", "", "config")
	rootCmd.AddCommand(runCmd)
}

func run() {
	var cfg config.Config
	b, err := base64.StdEncoding.DecodeString(configBase64)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cli := app.NewClient(&cfg)
	go cli.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
}

func runAsDaemon() {
	if p, running := isRunning(); running {
		fmt.Printf("%s is already running with PID: %d\n", programName, p.Pid)
		return
	}

	// 启动守护进程
	child, err := cntxt.Reborn()
	if err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}

	if child != nil {
		// 父进程，显示启动信息后退出
		fmt.Printf("Daemon started with PID: %d\n", child.Pid)
		fmt.Printf("Logs will be written to: %s\n", logFile)
		return
	}

	// 子进程继续运行
	defer cntxt.Release()
	run()
}
