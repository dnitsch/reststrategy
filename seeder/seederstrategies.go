package seeder

import (
	"context"
	"errors"
)

// GetPost strategy calls a GET endpoint and if item ***FOUND it does NOT do a POST***
// this strategy should be used sparingly and only in cases where the service REST implementation
// does not support an update of existing item.
func (r *SeederImpl) GetPost(ctx context.Context, action *Action) error {

	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	resp, err := r.get(ctx, action)
	if err != nil {
		return err
	}

	if string(resp) == "" {
		r.log.Infof("get endpoint returned no response, item can be created")
		return r.post(ctx, action)
	}
	r.log.Infof("get endpoint returned a response, item cannot be updated continuing")
	return nil
}

// FindPost strategy calls a GET endpoint and if item ***FOUND it does NOT do a POST***
// this strategy should be used sparingly and only in cases where the service REST implementation
// does not support an update of existing item.
func (r *SeederImpl) FindPost(ctx context.Context, action *Action) error {

	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	resp, err := r.get(ctx, action)
	if err != nil {
		return err
	}
	found, err := r.FindPathByExpression(resp, action.FindByJsonPathExpr)
	if err != nil {
		return err
	}
	// run Post as if its a PUT verb
	// when an id was found
	if found != "" && action.PostIsPut {
		return r.post(ctx, action)
	}

	if found == "" || action.PostIsPut {
		return r.post(ctx, action)
	}
	r.log.Infof("item: %s,found by expression: %s and cannot be updated continuing", found, action.FindByJsonPathExpr)
	return nil
}

// FindPutPost strategy gets an item either by specifying a known ID in the endpoint suffix
// or by pathExpression. Get can look for a response in an array or in a single response object.
// once a single item that matches is found and the relevant ID is extracted it will do a PUT
// else it will do a POST as the item can be created
func (r *SeederImpl) FindPutPost(ctx context.Context, action *Action) error {

	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	resp, err := r.get(ctx, action)
	if err != nil {
		return err
	}
	found, err := r.FindPathByExpression(resp, action.FindByJsonPathExpr)
	if err != nil {
		return err
	}
	if found == "" {
		r.log.Info("item not found by expression running POST")
		return r.post(ctx, action)
	}

	r.log.Infof("item: %s,found by expression: %s\nupdating in place", found, action.FindByJsonPathExpr)
	action.foundId = found
	return r.put(ctx, action)
}

// FindPatchPost is same as FindPutPost strategy but uses PATCH
func (r *SeederImpl) FindPatchPost(ctx context.Context, action *Action) error {

	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	resp, err := r.get(ctx, action)
	if err != nil {
		return err
	}
	found, err := r.FindPathByExpression(resp, action.FindByJsonPathExpr)
	if err != nil {
		return err
	}
	if found == "" {
		r.log.Info("item not found by expression running POST")
		return r.post(ctx, action)
	}

	r.log.Infof("item: %s,found by expression: %s\nupdating in place", found, action.FindByJsonPathExpr)
	action.foundId = found
	action.templatedPayload = r.TemplateWithVars(action.PatchPayloadTemplate, action.Variables)
	return r.patch(ctx, action)
}

// FindDeletePost
func (r *SeederImpl) FindDeletePost(ctx context.Context, action *Action) error {

	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	resp, err := r.get(ctx, action)
	if err != nil {
		var diag *Diagnostic
		if errors.As(err, &diag) {
			if !diag.ProceedFallback {
				return err
			}
			// proceeding with request
		} else {
			r.log.Debug("")
			return err
		}
	}
	found, err := r.FindPathByExpression(resp, action.FindByJsonPathExpr)
	if err != nil {
		return err
	}

	if found != "" {
		action.foundId = found
		r.log.Info("item found by expression running DELETE")
		if err := r.delete(ctx, action); err != nil {
			return err
		}
		return r.post(ctx, action)
	}

	r.log.Infof("item not found by expression: %s, creating...", found, action.FindByJsonPathExpr)
	return r.post(ctx, action)
}

// GetPutPost strategy gets an item by specifying a known ID in the endpoint suffix
// If a non error or non empty response is found it will do a PUT
// else it will do a POST as the item can be created
func (r *SeederImpl) GetPutPost(ctx context.Context, action *Action) error {

	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	resp, err := r.get(ctx, action)
	if err != nil {
		var diag *Diagnostic
		if errors.As(err, &diag) {
			if !diag.ProceedFallback {
				return err
			}
			// proceeding with request
		} else {
			r.log.Debug("")
			return err
		}
	}

	// // if FindByJsonPathExpr is provided we will try to extract it from response
	// // otherwise assume a known was supplied to the put endpoint
	// if action.FindByJsonPathExpr != "" && string(resp) == "" {
	// 	found, err := r.FindPathByExpression(resp, action.FindByJsonPathExpr)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	action.foundId = found
	// }

	if string(resp) == "" {
		r.log.Info("item not found. posting")
		return r.post(ctx, action)
	}

	r.log.Infof("found item: %v", string(resp))
	return r.put(ctx, action)
}

// Put strategy calls a PUT endpoint
// if standards compliant this should be an idempotent operation
func (r *SeederImpl) Put(ctx context.Context, action *Action) error {
	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	return r.put(ctx, action)
}

// Post strategy calls a POST endpoint
// This should not be an idempotent endpoint but in
// certain cases it is implemented with a 200 or 202.
func (r *SeederImpl) Post(ctx context.Context, action *Action) error {
	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	return r.post(ctx, action)
}

// Put strategy calls a PUT endpoint
// if standards compliant this should be an idempotent operation
func (r *SeederImpl) PutPost(ctx context.Context, action *Action) error {
	action.templatedPayload = r.TemplateWithVars(action.PayloadTemplate, action.Variables)
	if err := r.put(ctx, action); err != nil {
		var diag *Diagnostic
		if errors.As(err, &diag) {
			if diag.IsFatal || !diag.ProceedFallback {
				return err
			}
			r.log.Debug("falling back on POST")
			return r.post(ctx, action)
		}
	}
	return nil
}
