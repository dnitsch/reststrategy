package strategyrestseeder

import (
	"context"
	"io"
	"net/http"
	"os"

	log "github.com/dnitsch/simplelog"
	"github.com/dnitsch/strategyrestseeder/pkg/rest"
)

// StrategyConfig defines top level
type StrategyConfig struct {
	AuthConfig Auth                   `yaml:"auth"`
	Seeders    map[string]rest.Action `yaml:"seed"`
}

// Auth holds the auth strategy for
type Auth struct {
	AuthStrategy AuthType `yaml:"type"`
	username     string   `yaml:"username"`
	password     string   `yaml:"password"`
}

type StrategyFunc func(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error

type AuthType string

type StrategyType string

const (
	Basic        AuthType = "basicAuth"
	OAuth        AuthType = "oAuthClientCredentials"
	BasicToToken AuthType = "basicToToken"
)

const (
	GET_POST         StrategyType = "GET/POST"
	FIND_POST        StrategyType = "FIND/POST"
	POST_PUT         StrategyType = "POST/PUT"
	GET_PUT_POST     StrategyType = "GET/PUT/POST"
	FIND_PUT_POST    StrategyType = "FIND/PUT/POST"
	FIND_DELETE      StrategyType = "FIND/DELETE"
	FIND_DELETE_POST StrategyType = "FIND/DELETE/POST"
	PUT              StrategyType = "PUT"
)

type StrategyRestSeeder struct {
	Auth     Auth
	Strategy map[StrategyType]StrategyFunc
	rest     *rest.SeederImpl
	actions  []rest.Action
	log      log.Logger
}

// New initializes a default StrategySeeder with
// error log level and os.StdErr as log writer
// uses standard http.Client as rest client for rest SeederImplementation
func New() *StrategyRestSeeder {
	var gloglvl log.LogLevel = log.ErrorLvl
	l := log.New(os.Stderr, gloglvl)
	r := &rest.SeederImpl{}

	r.WithClient(&http.Client{}).WithLogger(l)

	return &StrategyRestSeeder{
		rest: r, // &rest.SeederImpl{client: &http.Client{}},
		Strategy: map[StrategyType]StrategyFunc{
			PUT:              PutStrategyFunc,
			POST_PUT:         PutPostStrategyFunc,
			GET_PUT_POST:     GetPutPostStrategyFunc,
			FIND_PUT_POST:    FindPutPostStrategyFunc,
			FIND_POST:        FindPostStrategyFunc,
			GET_POST:         GetPostStrategyFunc,
			FIND_DELETE_POST: FindDeletePostStrategyFunc,
		},
		log: l,
	}
}

// WithRestClient overwrites the default RestClient
func (s *StrategyRestSeeder) WithRestClient(rc rest.Client) *StrategyRestSeeder {
	s.rest = s.rest.WithClient(rc)
	return s
}

// WithLogger overwrites the default logger and passes it down to rest.SeederImpl
func (s *StrategyRestSeeder) WithLogger(w io.Writer, lvl log.LogLevel) *StrategyRestSeeder {
	s.log = log.New(w, lvl)
	s.rest = s.rest.WithLogger(s.log)
	return s
}

// WithActions builds the actions list
// empty actions will result in no restActions executing
func (s *StrategyRestSeeder) WithActions(actions map[string]rest.Action) *StrategyRestSeeder {
	for k, v := range actions {
		a := v
		s.actions = append(s.actions, *a.WithName(k))
	}
	return s
}

// Execute the built actions list
// TODO: create a custom error object
func (s *StrategyRestSeeder) Execute() []error {
	var errs []error
	// assign each action to method
	s.log.Debugf("actions: %v", s.actions)
	// TODO: when order is set
	// do some ordering if exists
	// else send to fixed size channel goroutine
	for _, action := range s.actions {
		if fn, ok := s.Strategy[StrategyType(action.Strategy)]; ok {
			e := fn(context.TODO(), &action, s.rest)
			if e != nil {
				errs = append(errs, e)
			}
		} else {
			s.log.Infof("unknown strategy")
		}
	}
	return errs
}

// PutStrategyFunc
func PutStrategyFunc(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error {
	return rest.Put(ctx, action)
}

// PutPostStrategyFunc known id to update
func PutPostStrategyFunc(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error {
	return rest.PutPost(ctx, action)
}

// FindPutPostStrategyFunc unknown ID and only know a name or other indicator
// the pathExpression must not evaluate to an empty string in order to for the PUT to be called
// else POST will be called as item was not present
func FindPutPostStrategyFunc(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error {
	return rest.FindPutPost(ctx, action)
}

// GetPutPostStrategyFunc known ID and only know a name or other indicator
// the pathExpression must not evaluate to an empty string in order to for the PUT to be called
// else POST will be called as item was not present
func GetPutPostStrategyFunc(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error {
	return rest.GetPutPost(ctx, action)
}

// FindDeletePostStrategyFunc
func FindDeletePostStrategyFunc(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error {
	return nil
}

// FindPostStrategyFunc strategy calls a GET endpoint and if item ***FOUND it does NOT do a POST***
// this strategy should be used sparingly and only in cases where the service REST implementation
// does not support an update of existing item.
func FindPostStrategyFunc(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error {
	return rest.FindPost(ctx, action)
}

// FindPostStrategyFunc strategy calls a GET endpoint and if item ***FOUND it does NOT do a POST***
// this strategy should be used sparingly and only in cases where the service REST implementation
// does not support an update of existing item.
func GetPostStrategyFunc(ctx context.Context, action *rest.Action, rest *rest.SeederImpl) error {
	return rest.GetPost(ctx, action)
}
