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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	repairmanv1 "github.com/awesomenix/repairman/api/v1"
)

// MaintenanceRequestReconciler reconciles a MaintenanceRequest object
type MaintenanceRequestReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancerequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancerequests/status,verbs=get;update;patch

func (r *MaintenanceRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("maintenancerequest", req.NamespacedName)

	mr := &repairmanv1.MaintenanceRequest{}
	err := r.Get(ctx, req.NamespacedName, mr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("received request", "name", req.Name, "type", mr.Spec.Type)
	ret := ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Minute}

	underMaintenance, err := r.getUnderMaintenanceCountByType(ctx, mr)
	if err != nil {
		log.Error(err, "failed to get under maintenance count by type", "type", mr.Spec.Type)
		return ret, nil
	}

	maintenanceLimit, err := r.getMaintenanceLimitsByType(ctx, mr)
	if err != nil {
		log.Error(err, "failed to get maintenance limits by type", "type", mr.Spec.Type)
		return ret, nil
	}

	if underMaintenance >= maintenanceLimit {
		log.Info("unabe to approve request, will exceed limits",
			"name", req.Name, "type", mr.Spec.Type,
			"um", underMaintenance, "ml", maintenanceLimit)
		return ret, nil
	}

	mr.Spec.State = "Pending"
	err = r.Get(ctx, req.NamespacedName, mr)
	if err != nil {
		log.Error(err, "failed to approve maintenance request", "name", req.Name, "type", mr.Spec.Type)
		return ret, nil
	}

	return ret, nil
}

func (r *MaintenanceRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&repairmanv1.MaintenanceRequest{}).
		Complete(r)
}

func (r *MaintenanceRequestReconciler) getUnderMaintenanceCountByType(ctx context.Context, mr *repairmanv1.MaintenanceRequest) (uint, error) {
	var umcount uint
	log := r.Log.WithValues("maintenancerequest", mr.Spec.Type)
	if !strings.EqualFold(mr.Spec.Type, "Node") {
		return umcount, errors.New("Unsupported maintenance type")
	}

	mrlist := &repairmanv1.MaintenanceRequestList{}
	err := r.List(ctx, mrlist)
	if err != nil {
		log.Error(err, "failed to list maintenance request", "type", mr.Spec.Type)
		return umcount, err
	}

	for _, mlr := range mrlist.Items {
		if !strings.EqualFold(mlr.Spec.Type, mr.Spec.Type) {
			continue
		}
		// if name is already under maintenance return error
		if strings.EqualFold(mlr.Spec.Name, mr.Spec.Name) {
			return umcount, errors.Errorf("already under maintenance, type: %s, name: %s", mr.Spec.Type, mlr.Spec.Name)
		}
		if mr.Spec.State == repairmanv1.Approved ||
			mr.Spec.State == repairmanv1.InProgress {
			log.Info("under maintenance", "type", mlr.Spec.Type, "state", mlr.Spec.State, "name", mlr.Name)
			umcount++
			continue
		}
		log.Info("not under maintenance", "type", mlr.Spec.Type, "state", mlr.Spec.State, "name", mlr.Name)
	}

	log.Info("total under maintenance count", "type", mr.Spec.Type, "count", umcount)
	return umcount, nil
}

func (r *MaintenanceRequestReconciler) getMaintenanceLimitsByType(ctx context.Context, mr *repairmanv1.MaintenanceRequest) (uint, error) {
	log := r.Log.WithValues("maintenancerequest", mr.Spec.Type)
	ml := &repairmanv1.MaintenanceLimit{}
	var mlcount uint
	err := r.Get(ctx, types.NamespacedName{Name: mr.Spec.Type}, ml)
	if err != nil {
		log.Error(err, "failed to get maintenance limit", "type", mr.Spec.Type)
		return mlcount, err
	}

	log.Info("maintenance limits", "type", mr.Spec.Type, "count", ml.Status.Limit)
	return ml.Spec.Limit, nil
}
