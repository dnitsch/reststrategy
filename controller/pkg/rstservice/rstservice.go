package rstservice

import (
	"context"

	"github.com/dnitsch/reststrategy/apis/reststrategy/v1alpha1"
	"github.com/dnitsch/reststrategy/seeder"
	"github.com/dnitsch/reststrategy/seeder/pkg/rest"
	log "github.com/dnitsch/simplelog"
)

type Seeder struct {
	log    log.Loggeriface
	seeder *seeder.StrategyRestSeeder
}

func New(log log.Loggeriface, rc rest.Client) *Seeder {
	srs := seeder.New(log).WithRestClient(rc)
	return &Seeder{
		log:    log,
		seeder: srs,
	}
}

// Execute accepts the top level spec
func (rst *Seeder) Execute(spec v1alpha1.StrategySpec) []error {
	rst.seeder.WithActions(rst.sortActions(spec.Seeders)).WithAuth(rst.sortAuth(spec.AuthConfig))
	return rst.seeder.Execute(context.Background())
}

func (rst *Seeder) sortActions(seeders []v1alpha1.SeederConfig) map[string]rest.Action {
	m := map[string]rest.Action{}
	for _, v := range seeders {
		m[v.Name] = v.Action
	}
	return m
}

func (rst *Seeder) sortAuth(auth []v1alpha1.AuthConfig) rest.AuthMap {
	m := rest.AuthMap{}
	for _, v := range auth {
		m[v.Name] = v.AuthConfig
	}
	return m
}

// create

// UpdateResource business implementation of whether to update the resource or not
func UpdateResource(new, old *v1alpha1.RestStrategy, doResync bool) bool {
	//
	return doResync || versionsDiffer(new, old)
}

// versionsDiffer returns true if new version is not in sync
func versionsDiffer(newDef, oldDef *v1alpha1.RestStrategy) bool {
	// check if there is a new version
	// be wary of CRDs applied via the SDK/pure API
	// may not correctly reflect Generation and ResourceVersion fields
	return oldDef.ObjectMeta.ResourceVersion != newDef.ObjectMeta.ResourceVersion && oldDef.ObjectMeta.Generation != newDef.ObjectMeta.Generation
}
