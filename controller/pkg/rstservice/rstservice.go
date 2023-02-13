package rstservice

import (
	"context"

	"github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
	"github.com/dnitsch/reststrategy/seeder"

	log "github.com/dnitsch/simplelog"
)

type Seeder struct {
	log    log.Loggeriface
	seeder *seeder.StrategyRestSeeder
}

func New(log log.Loggeriface, rc seeder.Client) *Seeder {
	srs := seeder.New(log).WithRestClient(rc)
	return &Seeder{
		log:    log,
		seeder: srs,
	}
}

// Execute accepts the top level spec
func (rst *Seeder) Execute(spec v1alpha1.StrategySpec) error {
	rst.seeder.WithActions(rst.sortActions(spec.Seeders)).WithAuth(rst.sortAuth(spec.AuthConfig))
	return rst.seeder.Execute(context.Background())
}

func (rst *Seeder) sortActions(seeders []v1alpha1.SeederConfig) map[string]seeder.Action {
	m := map[string]seeder.Action{}
	for _, v := range seeders {
		m[v.Name] = v.Action
	}
	return m
}

func (rst *Seeder) sortAuth(auth []v1alpha1.AuthConfig) seeder.AuthMap {
	m := seeder.AuthMap{}
	for _, v := range auth {
		m[v.Name] = v.AuthConfig
	}
	return m
}
