package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pocker",
	Short: "A Docker-like CLI wrapper around skopeo",
	Long: `pocker wraps skopeo with familiar Docker-style commands.

Examples:
  pocker pull ubuntu:latest
  pocker pull ghcr.io/username/image:tag --tag username/images:tag
  pocker save nginx:alpine /tmp/nginx.tar.gz
  pocker config all_proxy socks5://127.0.0.1:1080`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print pocker version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pocker %s\n", version)
	},
}
