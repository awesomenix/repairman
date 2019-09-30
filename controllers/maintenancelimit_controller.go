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
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	repairmanv1 "github.com/awesomenix/repairman/api/v1"
	corev1 "k8s.io/api/core/v1"
)

// PolicyHandler returns true if
type PolicyHandler func(corev1.Node) bool

// MaintenanceLimitReconciler reconciles a MaintenanceLimit object
type MaintenanceLimitReconciler struct {
	client.Client
	Log logr.Logger
	record.EventRecorder
}

// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancelimits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=repairman.k8s.io,resources=maintenancelimits/status,verbs=get;update;patch

// Reconcile ...
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

// SetupWithManager ...
func (r *MaintenanceLimitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&repairmanv1.MaintenanceLimit{}).
		Complete(r)
}

// UpdateMaintenanceLimits update maintenance limits if changed
func (r *MaintenanceLimitReconciler) UpdateMaintenanceLimits(ctx context.Context, log logr.Logger, ml *repairmanv1.MaintenanceLimit) error {
	if !strings.EqualFold(ml.Name, "node") {
		return errors.New("unsupported maintenance type")
	}
	var nodeList = &corev1.NodeList{}
	err := r.List(ctx, nodeList)
	if err != nil {
		log.Error(err, "failed to get", "type", ml.Name)
		return err
	}
	limit, unavailable := r.calculateMaintenanceLimit(ctx, ml, nodeList)

	if ml.Status.Limit != limit ||
		!reflect.DeepEqual(ml.Status.Unavailable, unavailable) {
		ml.Status.Limit = limit
		ml.Status.Unavailable = unavailable
		err = r.Status().Update(ctx, ml)
		if err != nil {
			return err
		}
		r.Eventf(ml, "Normal", "UpdatedLimits", "updated limits to %d of %d", limit, len(nodeList.Items))
	}
	return nil
}

// calculateMaintenanceLimit calculates the maximum number of nodes
// that can be updated at a given time
func (r *MaintenanceLimitReconciler) calculateMaintenanceLimit(ctx context.Context, ml *repairmanv1.MaintenanceLimit, nodeList *corev1.NodeList) (uint, map[string]repairmanv1.MaintenancePolicy) {
	if len(nodeList.Items) <= 0 {
		return 0, nil
	}

	unavailable := make(map[string]repairmanv1.MaintenancePolicy)
	for _, node := range nodeList.Items {
		if policy := r.nodeAvailable(node, ml); policy != repairmanv1.None {
			unavailable[node.Name] = policy
		}
	}

	nominalLimit := (int(ml.Spec.Limit) * len(nodeList.Items) / 100)
	calculatedLimit := math.Max(float64(nominalLimit), float64(1)) - float64(len(unavailable))
	return uint(math.Max(calculatedLimit, float64(0))), unavailable
}

func none(node corev1.Node) bool {
	return true
}

func notReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionFalse {
			return false
		}
	}
	return true
}

func unschedulable(node corev1.Node) bool {
	if node.Spec.Unschedulable {
		return false
	}
	return true
}

func policyHandler(policy repairmanv1.MaintenancePolicy) PolicyHandler {
	switch policy {
	case repairmanv1.NotReady:
		return notReady
	case repairmanv1.Unschedulable:
		return unschedulable
	}
	return none
}

// applyNodePolicies determines which nodes are effected by policies
func (r *MaintenanceLimitReconciler) nodeAvailable(node corev1.Node, ml *repairmanv1.MaintenanceLimit) repairmanv1.MaintenancePolicy {
	for _, policy := range ml.Spec.Policies {
		if !policyHandler(policy)(node) {
			return policy
		}
	}
	return repairmanv1.None
}
