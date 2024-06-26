package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/reststrategy/seeder/internal/config"
	"github.com/spf13/cobra"
)

var (
	Version               string
	Revision              string
	verbose               bool
	StrategyRestSeederCmd = &cobra.Command{
		Use:     config.SELF_NAME,
		Aliases: config.SHORT_NAME,
		Short:   fmt.Sprintf("%s CLI provides an idempotent rest caller", config.SELF_NAME),
		Long:    fmt.Sprintf(`%s CLI provides an idempotent rest caller for seeding configuration or data in a repeatable manner`, config.SELF_NAME),
		Version: fmt.Sprintf("%s-%s", Version, Revision),
	}
)

func Execute() {
	if err := StrategyRestSeederCmd.Execute(); err != nil {
		fmt.Errorf("cli error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	StrategyRestSeederCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbosity level")
}
