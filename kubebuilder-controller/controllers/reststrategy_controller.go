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
	"fmt"
	"net/http"
	"time"

	"github.com/dnitsch/reststrategy/kubebuilder-controller/api/v1alpha1"
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
	ResyncPeriod int64
}

//+kubebuilder:rbac:groups=seeder.dnitsch.net,resources=reststrategies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=seeder.dnitsch.net,resources=reststrategies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=seeder.dnitsch.net,resources=reststrategies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RestStrategy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *RestStrategyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	spec := v1alpha1.RestStrategy{}
	if err := r.Get(ctx, req.NamespacedName, &spec); err != nil {
		fmt.Println(err)
	}

	// restseeder
	restClient := &http.Client{}

	srs := seeder.New(r.Logger).WithRestClient(restClient)
	srs.WithActionsList(spec.Spec.Seeders).WithAuthFromList(spec.Spec.AuthConfig)
	if err := srs.Execute(context.Background()); err != nil {
		// update status as failed
		// requeue for 5mins
		return ctrl.Result{RequeueAfter: time.Duration(10) * time.Minute}, err
	}
	return ctrl.Result{RequeueAfter: time.Duration(r.ResyncPeriod) * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestStrategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&seederv1alpha1.RestStrategy{}).
		Complete(r)
}
