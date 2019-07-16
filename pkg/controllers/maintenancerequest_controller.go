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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
)

// MaintenanceRequestReconciler reconciles a MaintenanceRequest object
type MaintenanceRequestReconciler struct {
	client.Client
	Log logr.Logger
	record.EventRecorder
}

// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancerequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancerequests/status,verbs=get;update;patch

// Reconcile reconciles maintenance request
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
	return r.ApproveMaintenanceRequest(ctx, log, mr)
}

// ApproveMaintenanceRequest check if mr is under limits,
// approve if under limits
func (r *MaintenanceRequestReconciler) ApproveMaintenanceRequest(ctx context.Context, log logr.Logger, mr *repairmanv1.MaintenanceRequest) (ctrl.Result, error) {
	ret := ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Minute}

	if mr.Spec.State == repairmanv1.Approved ||
		mr.Spec.State == repairmanv1.Completed {
		return ret, nil
	}

	maintenanceLimit, err := r.getMaintenanceLimitsByType(ctx, mr)
	if err != nil {
		log.Error(err, "failed to get maintenance limits by type", "type", mr.Spec.Type)
		return ret, nil
	}

	underMaintenance, err := r.getUnderMaintenanceCountByType(ctx, mr)
	if err != nil {
		log.Error(err, "failed to get under maintenance count by type", "type", mr.Spec.Type)
		return ret, nil
	}

	if underMaintenance >= maintenanceLimit {
		log.Info("unable to approve request, will exceed limits",
			"name", mr.Name, "type", mr.Spec.Type,
			"um", underMaintenance, "ml", maintenanceLimit)
		return ret, nil
	}

	mr.Spec.State = "Approved"
	err = r.Update(ctx, mr)
	if err != nil {
		log.Error(err, "failed to approve maintenance request", "name", mr.Name, "type", mr.Spec.Type)
		return ret, nil
	}

	r.Eventf(mr, "Normal", "MaintenanceApproved", "approved maintenance request, under limits %d", maintenanceLimit)
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
	if !strings.EqualFold(mr.Spec.Type, "node") {
		return umcount, errors.New("Unsupported maintenance type")
	}

	mrlist := &repairmanv1.MaintenanceRequestList{}
	err := r.List(ctx, mrlist)
	if err != nil {
		log.Error(err, "failed to list maintenance requests", "type", mr.Spec.Type)
		return umcount, err
	}

	for _, mlr := range mrlist.Items {
		if !strings.EqualFold(mlr.Spec.Type, mr.Spec.Type) {
			continue
		}
		// if name is already under maintenance return error
		// if strings.EqualFold(mlr.Spec.Name, mr.Spec.Name) {
		// 	return umcount, errors.Errorf("already under maintenance, type: %s, name: %s", mr.Spec.Type, mlr.Spec.Name)
		// }
		if mlr.Spec.State == repairmanv1.Approved ||
			mlr.Spec.State == repairmanv1.InProgress {
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
		if apierrors.IsNotFound(err) {
			return math.MaxUint32, nil
		}
		log.Error(err, "failed to get maintenance limit", "type", mr.Spec.Type)
		return mlcount, err
	}

	log.Info("maintenance limits", "type", mr.Spec.Type, "count", ml.Status.Limit)
	return ml.Status.Limit, nil
}
