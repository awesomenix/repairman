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

// UpdateMaintenanceLimits update maintenance limits if changed
func (r *MaintenanceLimitReconciler) UpdateMaintenanceLimits(ctx context.Context, log logr.Logger, ml *repairmanv1.MaintenanceLimit) error {
	limit, err := r.getMaintenanceLimitsByType(ctx, log, ml)
	if err != nil {
		log.Error(err, "failed to get maintenance limits by type", "type", ml.Name)
		return err
	}
	var nodeList *corev1.NodeList
	err = r.getNodes(ctx, log, nodeList)
	if err != nil {
		log.Error(err, "failed to get maintenance limits by type", "type", ml.Name)
	}
	count := len(nodeList.Items)

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

// getMaintenanceLimitsByType calculates the max number of nodes that can be updated at the given time
func (r *MaintenanceLimitReconciler) getMaintenanceLimitsByType(ctx context.Context, log logr.Logger, ml *repairmanv1.MaintenanceLimit) (uint, error) {
	if !strings.EqualFold(ml.Name, "node") {
		return 0, errors.New("unsupported maintenance type")
	}

	var policied map[string]bool
	err := r.getPoliciedResourcesByType(ctx, log, ml, policied)
	if err != nil {
		log.Error(err, "unable to get policied resource", "type", ml.Name)
		return 0, err
	}

	var availableResources uint
	switch ml.Name {
	case "node":
		var nodeList *corev1.NodeList
		err := r.getNodes(ctx, log, nodeList)
		if err != nil {
			log.Error(err, "failed to list", "type", ml.Name)
			return 0, err
		}

		if len(nodeList.Items) <= 0 {
			return 0, nil
		}
		availableResources = r.calculateAvailableNodes(ctx, nodeList, policied)
	}

	nominalLimit := ml.Spec.Limit / 100 * availableResources
	return uint(math.Max(float64(nominalLimit), float64(1))), nil
}

func (r *MaintenanceLimitReconciler) getPoliciedResourcesByType(ctx context.Context, log logr.Logger, ml *repairmanv1.MaintenanceLimit, policied map[string]bool) error {
	switch ml.Name {
	case "node":
		var nodeList *corev1.NodeList
		err := r.getNodes(ctx, log, nodeList)
		if err != nil {
			log.Error(err, "failed to list", "type", ml.Name)
			return err
		}

		for _, node := range nodeList.Items {
			r.applyNodePolicies(node, ml, policied)
		}
	default:
		return errors.New("unsupported maintenance type")
	}

	return nil
}

func (r *MaintenanceLimitReconciler) applyNodePolicies(node corev1.Node, ml *repairmanv1.MaintenanceLimit, policied map[string]bool) {
	if ml.Spec.Policies.RespectUnschedulableNodes && node.Spec.Unschedulable {
		policied[node.Name] = true
	}

	if ml.Spec.Policies.RespectNotReadyNodes {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionFalse {
				policied[node.Name] = true
			}
		}
	}
}

func (r *MaintenanceLimitReconciler) calculateAvailableNodes(ctx context.Context, nodeList *corev1.NodeList, toRemove map[string]bool) uint {
	var available uint
	for _, node := range nodeList.Items {
		if _, ok := toRemove[node.Name]; ok {
			available++
		}
	}
	return available
}

func (r *MaintenanceLimitReconciler) getNodes(ctx context.Context, log logr.Logger, nodeList *corev1.NodeList) error {
	err := r.List(ctx, nodeList)
	if err != nil {
		log.Error(err, "failed to list nodes")
		return err
	}
	return nil
}
