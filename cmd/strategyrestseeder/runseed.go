package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	log "github.com/dnitsch/simplelog"
	srs "github.com/dnitsch/strategyrestseeder"
	"github.com/dnitsch/strategyrestseeder/internal/config"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

var (
	path   string
	runCmd = &cobra.Command{
		Use:     "run",
		Aliases: []string{"configure", "r"},
		Short:   fmt.Sprintf("Get version number %s", config.SELF_NAME),
		Long:    `Version and Revision number of the installed CLI`,
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
	strategyrestseederCmd.AddCommand(runCmd)
}

func runExecute(cmd *cobra.Command, args []string) error {

	strategy := srs.StrategyConfig{}
	_srs := srs.New().WithRestClient(&http.Client{})

	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	if err := yaml.Unmarshal(b, &strategy); err != nil {
		return err
	}
	_srs.WithActions(strategy.Seeders).WithAuth(&strategy.AuthConfig)

	if verbose {
		_srs.WithLogger(os.Stderr, log.DebugLvl)
	} else {
		_srs.WithLogger(os.Stderr, log.ErrorLvl)
	}

	if e := _srs.Execute(context.TODO()); len(e) > 0 {
		return fmt.Errorf("%+v", e)
	}
	return nil
}
