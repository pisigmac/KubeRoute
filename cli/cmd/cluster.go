package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/kuberoute/cli/pkg"
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage clusters",
}

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.Load()
		if err := pkg.RequireToken(cfg); err != nil {
			return err
		}
		var out struct {
			Clusters []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Status    string `json:"status"`
				NodeCount int    `json:"node_count"`
			} `json:"clusters"`
		}
		if err := pkg.NewClient(cfg).Do("GET", "/api/v1/clusters", nil, &out); err != nil {
			return err
		}
		fmt.Println(color.GreenString("ID") + "\t" + color.GreenString("NAME") + "\t" + color.GreenString("STATUS") + "\t" + color.GreenString("NODES"))
		for _, c := range out.Clusters {
			fmt.Printf("%s\t%s\t%s\t%d\n", c.ID, c.Name, c.Status, c.NodeCount)
		}
		return nil
	},
}

var clusterAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Register a new cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.Load()
		if err := pkg.RequireToken(cfg); err != nil {
			return err
		}
		provider, _ := cmd.Flags().GetString("provider")
		region, _ := cmd.Flags().GetString("region")
		var out map[string]any
		body := map[string]string{"name": args[0], "provider": provider, "region": region}
		if err := pkg.NewClient(cfg).Do("POST", "/api/v1/clusters", body, &out); err != nil {
			return err
		}
		fmt.Printf("Cluster registered id=%v\nSet KUBEROUTE_CLUSTER_ID to that id and run the agent.\n", out["id"])
		return nil
	},
}

var clusterStatusCmd = &cobra.Command{
	Use:   "status [id]",
	Short: "Show cluster health",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.Load()
		if err := pkg.RequireToken(cfg); err != nil {
			return err
		}
		var out map[string]any
		if err := pkg.NewClient(cfg).Do("GET", "/api/v1/clusters/"+args[0], nil, &out); err != nil {
			return err
		}
		fmt.Printf("Cluster: %v\nStatus: %v\nNodes: %v\nProvider: %v\n", out["name"], out["status"], out["node_count"], out["provider"])
		return nil
	},
}

func init() {
	clusterAddCmd.Flags().String("provider", "kind", "provider (kind, aws, gcp, azure, on-prem)")
	clusterAddCmd.Flags().String("region", "local", "region")
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterAddCmd)
	clusterCmd.AddCommand(clusterStatusCmd)
}
