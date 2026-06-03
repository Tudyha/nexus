package cmd

import (
	"fmt"

	"github.com/Tudyha/nexus/client/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "查看客户端版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Nexus CLI %s (build %d)\n", version.VersionName, version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
