package cmd

import (
	"fmt"

	"github.com/kuberoute/cli/pkg"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Kuberoute",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.Load()
		client := pkg.NewClient(cfg)
		var out struct {
			Token string `json:"token"`
		}
		if err := client.Do("GET", "/auth/dev?format=json", nil, &out); err != nil {
			return fmt.Errorf("login failed (local dev auth): %w", err)
		}
		cfg.Token = out.Token
		if err := pkg.Save(cfg); err != nil {
			return err
		}
		fmt.Println("Logged in. Token stored in ~/.kuberoute/config.yaml")
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kr v0.2.0-dev")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(versionCmd)
}
