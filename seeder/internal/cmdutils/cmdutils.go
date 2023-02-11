package cmdutils

import (
	"context"
	"fmt"

	srs "github.com/dnitsch/reststrategy/seeder"
)

func RunSeed(svc *srs.StrategyRestSeeder, strategy srs.StrategyConfig, path string, verbose bool) error {

	svc.WithActions(strategy.Seeders).WithAuth(strategy.AuthConfig)

	if e := svc.Execute(context.TODO()); e != nil {
		return fmt.Errorf("%+v", e)
	}
	return nil
}
