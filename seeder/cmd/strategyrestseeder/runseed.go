package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
	"gopkg.in/yaml.v3"

	srs "github.com/dnitsch/reststrategy/seeder"
	"github.com/dnitsch/reststrategy/seeder/internal/cmdutils"
	"github.com/dnitsch/reststrategy/seeder/internal/config"
	log "github.com/dnitsch/simplelog"
	"github.com/spf13/cobra"
)

var (
	path                string
	enableConfigManager bool
	cmTokenSeparator    string
	cmKeySeparator      string
	runCmd              = &cobra.Command{
		Use:     "run",
		Aliases: config.SHORT_NAME,
		Short:   `Executes the provided strategy`,
		Long:    `Executes the provided strategy against the provided actions and auth references`,
		RunE:    runExecute,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(path) < 1 {
				return fmt.Errorf("must include input for path")
			}
			return nil
		},
	}
)

func init() {
	runCmd.PersistentFlags().StringVarP(&path, "path", "p", "", `Path to YAML file which has the strategy defined`)
	runCmd.PersistentFlags().BoolVarP(&enableConfigManager, "enable-config-manager", "c", false, "Enables config manager to replace placeholders for secret values")
	runCmd.PersistentFlags().StringVarP(&cmTokenSeparator, "cm-token-separator", "t", "", `Config Manager token separator`)
	runCmd.PersistentFlags().StringVarP(&cmKeySeparator, "cm-key-separator", "k", "", `Config Manager key separator`)
	StrategyRestSeederCmd.AddCommand(runCmd)
}

func runExecute(cmd *cobra.Command, args []string) error {

	l := log.New(os.Stderr, log.ErrorLvl)

	if verbose {
		l = log.New(os.Stderr, log.DebugLvl)
	}

	strategy := &srs.StrategyConfig{}
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(file, strategy); err != nil {
		return err
	}

	s := srs.New(&l).WithRestClient(&http.Client{})
	cmConfig := generator.NewConfig()

	if cmKeySeparator != "" {
		cmConfig.WithKeySeparator(cmKeySeparator)
	}

	if cmTokenSeparator != "" {
		cmConfig.WithTokenSeparator(cmTokenSeparator)
	}

	if enableConfigManager {
		s.WithConfigManager(&configmanager.ConfigManager{}).WithConfigManagerOptions(cmConfig)
	}

	return cmdutils.RunSeed(s, strategy)
}
