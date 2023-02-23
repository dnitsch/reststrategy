/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"net/http"
	"time"

	seederv1alpha1 "github.com/dnitsch/reststrategy/kubebuilder-controller/api/v1alpha1"
	"github.com/dnitsch/reststrategy/seeder"
	log "github.com/dnitsch/simplelog"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RestStrategyReconciler reconciles a RestStrategy object
type RestStrategyReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	Logger       log.Loggeriface
	IsNamespaced bool
	Namespace    string
	// ResyncPeriod in minutes
	// the amount of minutes to wait after successful
	// apply of resource before re-applying again
	//
	ResyncPeriod int
	// FailedResourceResyncPeriod the amount of time
	// to wait after a failed item was processed
	FailedResourceResyncPeriod int
}

func (r *RestStrategyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	spec := seederv1alpha1.RestStrategy{}
	if err := r.Get(ctx, req.NamespacedName, &spec); err != nil {
		r.Logger.Errorf("failed to get resource: %v", err.Error())
	}

	// restseeder
	restClient := &http.Client{}

	srs := seeder.New(r.Logger).WithRestClient(restClient)
	srs.WithActionsList(spec.Spec.Seeders).WithAuthFromList(spec.Spec.AuthConfig)
	if err := srs.Execute(context.Background()); err != nil {
		spec.Status.Message = err.Error()
		// update status as failed
		// requeue for 5mins
		_ = r.Status().Update(ctx, &spec)
		return ctrl.Result{RequeueAfter: time.Duration(10) * time.Minute}, err
	}
	spec.Status.Message = "Sucessfully Synced"
	_ = r.Status().Update(ctx, &spec)
	return ctrl.Result{RequeueAfter: time.Duration(r.ResyncPeriod) * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestStrategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&seederv1alpha1.RestStrategy{}).
		Complete(r)
}
