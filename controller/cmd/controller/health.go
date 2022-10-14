package controller

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	healthCmd        = &cobra.Command{
		Use:   "health",
		Short: fmt.Sprintf("Check app health"),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("HealthCheckRequestd")
			os.Exit(0)
		},
	}
)

func init() {
	controllerCmd.AddCommand(healthCmd)
}
