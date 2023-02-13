package seeder

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	log "github.com/dnitsch/simplelog"
)

type StrategyFunc func(ctx context.Context, action *Action, rest *SeederImpl) error

type StrategyType string

const (
	GET_POST         StrategyType = "GET/POST"
	FIND_POST        StrategyType = "FIND/POST"
	PUT_POST         StrategyType = "PUT/POST"
	GET_PUT_POST     StrategyType = "GET/PUT/POST"
	FIND_PUT_POST    StrategyType = "FIND/PUT/POST"
	FIND_PATCH_POST  StrategyType = "FIND/PATCH/POST"
	FIND_DELETE      StrategyType = "FIND/DELETE"
	FIND_DELETE_POST StrategyType = "FIND/DELETE/POST"
	PUT              StrategyType = "PUT"
)

type StrategyRestSeeder struct {
	Strategy map[StrategyType]StrategyFunc
	rest     *SeederImpl
	actions  []Action
	log      log.Loggeriface
}

// New initializes a default StrategySeeder with
// error log level and os.StdErr as log writer
// uses standard http.Client as rest client for rest SeederImplementation
func New(log log.Loggeriface) *StrategyRestSeeder {
	r := NewSeederImpl(log)
	r.WithClient(&http.Client{})

	return &StrategyRestSeeder{
		rest: r,
		Strategy: map[StrategyType]StrategyFunc{
			PUT:              PutStrategyFunc,
			PUT_POST:         PutPostStrategyFunc,
			GET_PUT_POST:     GetPutPostStrategyFunc,
			FIND_PUT_POST:    FindPutPostStrategyFunc,
			FIND_POST:        FindPostStrategyFunc,
			GET_POST:         GetPostStrategyFunc,
			FIND_DELETE_POST: FindDeletePostStrategyFunc,
			FIND_PATCH_POST:  FindPatchPostStrategyFunc,
		},
		log: log,
	}
}

// WithRestClient overwrites the default RestClient
func (s *StrategyRestSeeder) WithRestClient(rc Client) *StrategyRestSeeder {
	s.rest = s.rest.WithClient(rc)
	return s
}

// WithAuth adds the AuthLogic to the entire seeder
// NOTE: might make more sense to have a per RestAction authTemplate (might make it very inefficient)
func (s *StrategyRestSeeder) WithAuth(ra AuthMap) *StrategyRestSeeder {
	s.rest = s.rest.WithAuth(ra)
	return s
}

// WithActions builds the actions list
// empty actions will result in no restActions executing
func (s *StrategyRestSeeder) WithActions(actions map[string]Action) *StrategyRestSeeder {
	for k, v := range actions {
		a := v
		s.actions = append(s.actions, *a.WithName(k).WithHeader())
	}
	return s
}

// Execute the built actions list
func (s *StrategyRestSeeder) Execute(ctx context.Context) error {
	var errs []error
	// assign each action to method
	s.log.Debugf("actions: %v", s.actions)
	// do some ordering if exists
	// else send to fixed size channel goroutine
	for _, action := range s.actions {
		if fn, ok := s.Strategy[StrategyType(action.Strategy)]; ok {
			e := fn(ctx, &action, s.rest)
			if e != nil {
				errs = append(errs, e)
				continue
			}
			continue
		}
		s.log.Infof("unknown strategy")
	}
	if len(errs) > 0 {
		finalErr := []string{}
		for _, e := range errs {
			finalErr = append(finalErr, e.Error())
		}
		return fmt.Errorf(strings.Join(finalErr, "\n"))
	}
	return nil
}

// PutStrategyFunc calls a PUT endpoint fails if an error occurs
// useful when there is a known Id of a resource and PUT supports creation
func PutStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.Put(ctx, action)
}

// PutPostStrategyFunc is useful when the resource is created a user specified Id
// the PUT endpoint DOES NOT support a creation of the resource. PUT should throw a 4XX
// for the POST fallback to take effect
func PutPostStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.PutPost(ctx, action)
}

// FindPutPostStrategyFunc is useful when the resource Id is unknown i.e. handled by the system.
// providing a pathExpression will evaluate the response.
// the pathExpression must not evaluate to an empty string in order to for the PUT to be called
// else POST will be called as item was not present
func FindPutPostStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.FindPutPost(ctx, action)
}

// FindPatchPostStrategyFunc same as FindPutPostStrategyFunc but uses PATCH instead of PUT
func FindPatchPostStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.FindPatchPost(ctx, action)
}

// GetPutPostStrategyFunc known ID and only know a name or other indicator
// the pathExpression must not evaluate to an empty string in order to for the PUT to be called
// else POST will be called as item was not present
func GetPutPostStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.GetPutPost(ctx, action)
}

// FindDeletePostStrategyFunc is useful for when you cannot update a resource
// but it can be safely destroyed an recreated
func FindDeletePostStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.FindDeletePost(ctx, action)
}

// FindPostStrategyFunc strategy calls a GET endpoint and expects a list of items to match against a
// FilterPathExpre and if item ***FOUND*** it does _NOT_ do a ***POST***
// this strategy should be used sparingly and only in cases where the service REST implementation
// does not support an update of existing item.
func FindPostStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.FindPost(ctx, action)
}

// GetPostStrategyFunc strategy calls a GET endpoint by an ID and if item ***FOUND*** it does _NOT_ do a ***POST***
// this strategy should be used sparingly and only in cases where the service REST implementation
// does not support an update of existing item.
func GetPostStrategyFunc(ctx context.Context, action *Action, rest *SeederImpl) error {
	return rest.GetPost(ctx, action)
}
