package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/kuberoute/cli/pkg"
	"github.com/spf13/cobra"
)

var decisionsCmd = &cobra.Command{
	Use:   "decisions",
	Short: "Manage routing decisions",
}

var clusterIDFlag string

var decisionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending decisions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.Load()
		if err := pkg.RequireToken(cfg); err != nil {
			return err
		}
		id, err := requireCluster(cfg)
		if err != nil {
			return err
		}
		var out struct {
			Decisions []struct {
				ID             string  `json:"id"`
				DecisionType   string  `json:"decision_type"`
				ConfidenceScore float64 `json:"confidence_score"`
				Reason         string  `json:"reason"`
				Status         string  `json:"status"`
			} `json:"decisions"`
		}
		if err := pkg.NewClient(cfg).Do("GET", "/api/v1/clusters/"+id+"/decisions", nil, &out); err != nil {
			return err
		}
		fmt.Println(color.YellowString("ID") + "\t" + color.YellowString("TYPE") + "\t" + color.YellowString("CONF") + "\t" + color.YellowString("STATUS") + "\t" + color.YellowString("REASON"))
		for _, d := range out.Decisions {
			fmt.Printf("%s\t%s\t%.2f\t%s\t%s\n", d.ID, d.DecisionType, d.ConfidenceScore, d.Status, d.Reason)
		}
		return nil
	},
}

var decisionsApplyCmd = &cobra.Command{
	Use:   "apply [decision-id]",
	Short: "Apply a decision",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return decisionAction(args[0], "apply")
	},
}

var decisionsRejectCmd = &cobra.Command{
	Use:   "reject [decision-id]",
	Short: "Reject a decision",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return decisionAction(args[0], "reject")
	},
}

var decisionsShowCmd = &cobra.Command{
	Use:   "show [decision-id]",
	Short: "Show a placement PR (solver trace, patch, timeline)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := pkg.Load()
		if err := pkg.RequireToken(cfg); err != nil {
			return err
		}
		id, err := requireCluster(cfg)
		if err != nil {
			return err
		}
		var out map[string]any
		if err := pkg.NewClient(cfg).Do("GET", "/api/v1/clusters/"+id+"/decisions/"+args[0], nil, &out); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var decisionsRevertCmd = &cobra.Command{
	Use:   "revert [decision-id]",
	Short: "Open a revert placement for an applied decision",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return decisionAction(args[0], "revert")
	},
}

func decisionAction(decisionID, action string) error {
	cfg := pkg.Load()
	if err := pkg.RequireToken(cfg); err != nil {
		return err
	}
	id, err := requireCluster(cfg)
	if err != nil {
		return err
	}
	if err := pkg.NewClient(cfg).Do("POST", "/api/v1/clusters/"+id+"/decisions/"+decisionID+"/"+action, nil, nil); err != nil {
		return err
	}
	fmt.Printf("Decision %s %s.\n", decisionID, action)
	return nil
}

func requireCluster(cfg pkg.Config) (string, error) {
	if clusterIDFlag != "" {
		return clusterIDFlag, nil
	}
	var out struct {
		Clusters []struct {
			ID string `json:"id"`
		} `json:"clusters"`
	}
	if err := pkg.NewClient(cfg).Do("GET", "/api/v1/clusters", nil, &out); err != nil {
		return "", err
	}
	if len(out.Clusters) == 0 {
		return "", fmt.Errorf("no clusters; pass --cluster")
	}
	return out.Clusters[0].ID, nil
}

func init() {
	decisionsCmd.PersistentFlags().StringVar(&clusterIDFlag, "cluster", "", "cluster id")
	rootCmd.AddCommand(decisionsCmd)
	decisionsCmd.AddCommand(decisionsListCmd)
	decisionsCmd.AddCommand(decisionsShowCmd)
	decisionsCmd.AddCommand(decisionsApplyCmd)
	decisionsCmd.AddCommand(decisionsRejectCmd)
	decisionsCmd.AddCommand(decisionsRevertCmd)
}
