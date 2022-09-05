package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/strategyrestseeder/internal/config"
	"github.com/spf13/cobra"
)

var (
	verbose               bool
	tokenSeparator        string
	keySeparator          string
	strategyrestseederCmd = &cobra.Command{
		Use:     config.SELF_NAME,
		Aliases: config.SHORT_NAME,
		Short:   fmt.Sprintf("%s CLI for retrieving and inserting config or secret variables", config.SELF_NAME),
		Long: fmt.Sprintf(`%s CLI for retrieving config or secret variables.
		Using a specific tokens as an array item`, config.SELF_NAME),
	}
)

func Execute() {
	if err := strategyrestseederCmd.Execute(); err != nil {
		fmt.Errorf("cli error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	strategyrestseederCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbosity level")
}
