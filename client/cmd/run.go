package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tudyha/nexus/client/app"
	"github.com/Tudyha/nexus/client/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	daemonFlag   bool
	configBase64 string
	configFile   string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "启动客户端连接服务端",
	Long: `启动 Nexus 客户端并连接到服务端。

连接成功后客户端会与服务端建立 smux 多路复用连接，等待服务端下发隧道任务。
客户端会定期发送心跳以维持连接，并在断连时自动重连。

配置文件的 JSON 格式:
{
  "server_addr": "192.168.1.100:8081",
  "app_id": 1,
  "app_secret": "your-app-secret",
  "heartbeat_interval": 30,
  "connect_timeout": 10,
  "reconnect_interval": 5
}

选项:
  -f, --config-file    从 JSON 文件读取配置
  -c, --config         从 Base64 编码的 JSON 字符串读取配置
  -d, --daemon         以后台守护进程模式运行

示例:
  nexus-cli run -f config.json
  nexus-cli run -d -f config.json
  nexus-cli run -c $(base64 -w0 config.json)
`,
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
	runCmd.Flags().StringVarP(&configBase64, "config", "c", "", "Base64 encoded config JSON")
	runCmd.Flags().StringVarP(&configFile, "config-file", "f", "", "Path to config JSON file")
	rootCmd.AddCommand(runCmd)
}

func run() {
	var cfg config.Config
	b, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "config json unmarshal failed: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	cli, err := app.NewClient(ctx, &cfg, app.Options{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		cli.Shutdown()
		cancel()
	}()

	if err := cli.Run(); err != nil && err != context.Canceled {
		log.Error().Err(err).Msg("client run error")
		os.Exit(1)
	}
	log.Info().Msg("client stopped gracefully")
}

func loadConfig() ([]byte, error) {
	if configFile != "" {
		return os.ReadFile(configFile)
	}
	if configBase64 != "" {
		return base64.StdEncoding.DecodeString(configBase64)
	}
	return nil, fmt.Errorf("either --config (-c) or --config-file (-f) is required")
}

func runAsDaemon() {
	if p, running := isRunning(); running {
		fmt.Printf("%s is already running with PID: %d\n", programName, p.Pid)
		return
	}

	// 启动守护进程
	child, err := cntxt.Reborn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
		os.Exit(1)
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
