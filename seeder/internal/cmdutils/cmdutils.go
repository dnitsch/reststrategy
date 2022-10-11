package cmdutils

import (
	"context"
	"fmt"
	"os"

	srs "github.com/dnitsch/reststrategy/seeder"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"
)

func RunSeed(svc *srs.StrategyRestSeeder, strategy rest.StrategyConfig, path string, verbose bool) error {
	var l log.Logger

	if verbose {
		l = log.New(os.Stderr, log.DebugLvl)
	} else {
		l = log.New(os.Stderr, log.ErrorLvl)
	}

	svc.WithActions(strategy.Seeders).WithAuth(&strategy.AuthConfig)
	svc.WithLogger(l)

	if e := svc.Execute(context.TODO()); len(e) > 0 {
		return fmt.Errorf("%+v", e)
	}
	return nil
}
