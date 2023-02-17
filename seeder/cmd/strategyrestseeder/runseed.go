package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/reststrategy/seeder"
	srs "github.com/dnitsch/reststrategy/seeder"
	"github.com/dnitsch/reststrategy/seeder/internal/cmdutils"
	"github.com/dnitsch/reststrategy/seeder/internal/config"
	log "github.com/dnitsch/simplelog"
	"github.com/spf13/cobra"
)

var (
	path   string
	runCmd = &cobra.Command{
		Use:     "run",
		Aliases: config.SHORT_NAME,
		Short:   `Executes the provided strategy`,
		Long:    `Executes the provided strategy against the provided actions and auth references`,
		RunE:    runExecute,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// if len(input) < 1 && !getStdIn() {
			if len(path) < 1 {
				return fmt.Errorf("must include input")
			}
			return nil
		},
	}
)

func init() {
	runCmd.PersistentFlags().StringVarP(&path, "path", "p", "", `Path to YAML file which has the strategy defined`)
	// expose configmanager enabled flag here
	strategyrestseederCmd.AddCommand(runCmd)
}

func runExecute(cmd *cobra.Command, args []string) error {

	var l log.Logger

	if verbose {
		l = log.New(os.Stderr, log.DebugLvl)
	} else {
		l = log.New(os.Stderr, log.ErrorLvl)
	}

	strategy := seeder.StrategyConfig{}
	s := srs.New(&l).WithRestClient(&http.Client{})

	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}

	cm := &configmanager.ConfigManager{}
	config := generator.NewConfig().WithTokenSeparator("://")
	if _, err := configmanager.RetrieveUnmarshalledFromYaml(b, &strategy, cm, *config); err != nil {
		return err
	}

	return cmdutils.RunSeed(s, strategy, path, verbose)
}
