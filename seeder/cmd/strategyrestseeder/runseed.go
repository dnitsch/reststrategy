package cmd

import (
	"fmt"
	"net/http"
	"os"

	srs "github.com/dnitsch/reststrategy/seeder"
	"github.com/dnitsch/reststrategy/seeder/internal/cmdutils"
	"github.com/dnitsch/reststrategy/seeder/internal/config"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"
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

	var l log.Logger

	if verbose {
		l = log.New(os.Stderr, log.DebugLvl)
	} else {
		l = log.New(os.Stderr, log.ErrorLvl)
	}

	strategy := rest.StrategyConfig{}
	s := srs.New(&l).WithRestClient(&http.Client{})

	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}

	if err := yaml.Unmarshal(b, &strategy); err != nil {
		return err
	}

	return cmdutils.RunSeed(s, strategy, path, verbose)
}
