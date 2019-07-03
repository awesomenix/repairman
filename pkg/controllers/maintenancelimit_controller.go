/*

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
	"math"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

// MaintenanceLimitReconciler reconciles a MaintenanceLimit object
type MaintenanceLimitReconciler struct {
	client.Client
	Log logr.Logger
	record.EventRecorder
}

// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancelimits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancelimits/status,verbs=get;update;patch

func (r *MaintenanceLimitReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("maintenancelimit", req.NamespacedName)

	ml := &repairmanv1.MaintenanceLimit{}
	err := r.Get(ctx, req.NamespacedName, ml)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("received request", "name", req.Name, "limit", ml.Spec.Limit)
	err = r.UpdateMaintenanceLimits(ctx, log, ml)
	if err != nil {
		log.Error(err, "failed to get maintenance limits by type", "type", ml.Name)
	}
	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

func (r *MaintenanceLimitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&repairmanv1.MaintenanceLimit{}).
		Complete(r)
}

func (r *MaintenanceLimitReconciler) getMaintenanceLimitsByType(ctx context.Context, ml *repairmanv1.MaintenanceLimit) (uint, uint, error) {
	var limit uint
	log := r.Log.WithValues("maintenancelimit", ml.Name)
	if !strings.EqualFold(ml.Name, "node") {
		return limit, limit, errors.New("unsupported maintenance type")
	}

	nodelist := &corev1.NodeList{}
	err := r.List(ctx, nodelist)
	if err != nil {
		log.Error(err, "failed to list", "type", ml.Name)
		return limit, limit, err
	}

	if len(nodelist.Items) <= 0 {
		return limit, limit, nil
	}

	limit = (ml.Spec.Limit * uint(len(nodelist.Items))) / 100
	limit = uint(math.Max(float64(limit), float64(1)))
	return limit, uint(len(nodelist.Items)), nil
}

// UpdateMaintenanceLimits update maintenance limits if changed
func (r *MaintenanceLimitReconciler) UpdateMaintenanceLimits(ctx context.Context, log logr.Logger, ml *repairmanv1.MaintenanceLimit) error {
	limit, count, err := r.getMaintenanceLimitsByType(ctx, ml)
	if err != nil {
		log.Error(err, "failed to get maintenance limits by type", "type", ml.Name)
		return err
	}

	if ml.Status.Limit != limit {
		ml.Status.Limit = limit
		err = r.Status().Update(ctx, ml)
		if err != nil {
			return err
		}
		r.Eventf(ml, "Normal", "UpdatedLimits", "updated limits to %d of %d", limit, count)
	}

	return nil
}
