package cmd

import (
	"fmt"

	"github.com/kuberoute/cli/pkg"
	"github.com/spf13/cobra"
)

var costsCmd = &cobra.Command{
	Use:   "costs",
	Short: "Cost reports and savings",
}

var costsReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show cost savings",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.Load()
		if err := pkg.RequireToken(cfg); err != nil {
			return err
		}
		var out struct {
			Total     float64 `json:"total"`
			Optimized float64 `json:"optimized"`
			Savings   float64 `json:"savings"`
		}
		if err := pkg.NewClient(cfg).Do("GET", "/api/v1/costs", nil, &out); err != nil {
			return err
		}
		fmt.Println("Cost Report (from cluster snapshot + applied decisions)")
		fmt.Printf("Total: $%.2f\n", out.Total)
		fmt.Printf("Optimized: $%.2f\n", out.Optimized)
		fmt.Printf("Savings: $%.2f\n", out.Savings)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(costsCmd)
	costsCmd.AddCommand(costsReportCmd)
}
