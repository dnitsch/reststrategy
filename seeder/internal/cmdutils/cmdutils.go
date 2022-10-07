package cmdutils

import (
	"context"
	"fmt"
	"os"

	srs "github.com/dnitsch/reststrategy/seeder"
	log "github.com/dnitsch/simplelog"
)

func RunSeed(svc *srs.StrategyRestSeeder, strategy srs.StrategyConfig, path string, verbose bool) error {

	svc.WithActions(strategy.Seeders).WithAuth(&strategy.AuthConfig)
	svc.WithLogger(os.Stderr, log.ErrorLvl)

	if verbose {
		svc.WithLogger(os.Stderr, log.DebugLvl)
	}

	if e := svc.Execute(context.TODO()); len(e) > 0 {
		return fmt.Errorf("%+v", e)
	}
	return nil
}
