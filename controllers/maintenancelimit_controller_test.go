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
	"fmt"
	"testing"

	repairmanv1 "github.com/awesomenix/repairman/api/v1"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testSetup(t *testing.T) (*assert.Assertions, MaintenanceLimitReconciler) {
	assert := assert.New(t)
	repairmanv1.AddToScheme(scheme.Scheme)
	f := fake.NewFakeClient()

	return assert, MaintenanceLimitReconciler{
		Client:        f,
		Log:           ctrl.Log,
		EventRecorder: &record.FakeRecorder{},
	}
}

func TestUpdateMaitenanceLimits(t *testing.T) {
	assert := assert.New(t)
	repairmanv1.AddToScheme(scheme.Scheme)
	f := fake.NewFakeClient()

	reconciler := &MaintenanceLimitReconciler{
		Client:        f,
		EventRecorder: &record.FakeRecorder{},
		Log:           ctrl.Log,
	}

	for i := 0; i < 1; i++ {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("dummynode%d", i),
			},
		}
		err := f.Create(context.TODO(), node)
		assert.Nil(err)
	}

	err := reconciler.UpdateMaintenanceLimits(context.TODO(), reconciler.Log, &repairmanv1.MaintenanceLimit{})
	assert.NotNil(err)
	assert.Contains(err.Error(), "unsupported maintenance type")

	ml := &repairmanv1.MaintenanceLimit{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node",
		},
		Spec: repairmanv1.MaintenanceLimitSpec{
			Limit: 50,
		},
	}
	err = f.Create(context.TODO(), ml)
	err = reconciler.UpdateMaintenanceLimits(context.TODO(), reconciler.Log, ml)
	assert.Nil(err)
	assert.Equal(uint(1), ml.Status.Limit)

	for i := 1; i < 10; i++ {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("dummynode%d", i),
			},
		}
		err := f.Create(context.TODO(), node)
		assert.Nil(err)
	}
	err = reconciler.UpdateMaintenanceLimits(context.TODO(), reconciler.Log, ml)
	assert.Nil(err)
	assert.Equal(uint(5), ml.Status.Limit)
}

func TestCalculateMaintenanceLimit(t *testing.T) {
	var tests = []struct {
		maintenanceLimit *repairmanv1.MaintenanceLimit
		nodeList         *corev1.NodeList
		limit            uint
	}{
		{
			&repairmanv1.MaintenanceLimit{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
				Spec: repairmanv1.MaintenanceLimitSpec{
					Limit: 50,
				},
			}, &corev1.NodeList{}, 0,
		},
		{
			&repairmanv1.MaintenanceLimit{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
				Spec: repairmanv1.MaintenanceLimitSpec{
					Limit: 50,
				},
			}, &corev1.NodeList{
				Items: []corev1.Node{
					corev1.Node{},
					corev1.Node{},
					corev1.Node{},
					corev1.Node{},
				},
			}, 2,
		},
		{
			&repairmanv1.MaintenanceLimit{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
				Spec: repairmanv1.MaintenanceLimitSpec{
					Limit: 50,
				},
			}, &corev1.NodeList{
				Items: []corev1.Node{
					corev1.Node{},
					corev1.Node{},
				},
			}, 1,
		},
		{
			&repairmanv1.MaintenanceLimit{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
				Spec: repairmanv1.MaintenanceLimitSpec{
					Limit:    50,
					Policies: []repairmanv1.MaintenancePolicy{repairmanv1.Unschedulable},
				},
			}, &corev1.NodeList{
				Items: []corev1.Node{
					corev1.Node{},
					corev1.Node{
						Spec: corev1.NodeSpec{
							Unschedulable: true,
						},
					},
				},
			}, 0,
		},
		{
			&repairmanv1.MaintenanceLimit{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
				Spec: repairmanv1.MaintenanceLimitSpec{
					Limit:    50,
					Policies: []repairmanv1.MaintenancePolicy{repairmanv1.NotReady},
				},
			}, &corev1.NodeList{
				Items: []corev1.Node{
					corev1.Node{},
					corev1.Node{
						Spec: corev1.NodeSpec{
							Unschedulable: true,
						},
					},
				},
			}, 1,
		},
	}
	assert, reconciler := testSetup(t)
	for _, tt := range tests {
		limit := reconciler.calculateMaintenanceLimit(context.TODO(), tt.maintenanceLimit, tt.nodeList)
		assert.Equal(tt.limit, limit)
	}
}

func TestApplyNodePolicy(t *testing.T) {
	var tests = []struct {
		node             corev1.Node
		maintenanceLimit *repairmanv1.MaintenanceLimit
		expected         bool
	}{
		{
			corev1.Node{
				Spec: corev1.NodeSpec{
					Unschedulable: true,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
			},
			&repairmanv1.MaintenanceLimit{
				Spec: repairmanv1.MaintenanceLimitSpec{},
			},
			true,
		},
		{
			corev1.Node{
				Spec: corev1.NodeSpec{
					Unschedulable: true,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
			},
			&repairmanv1.MaintenanceLimit{
				Spec: repairmanv1.MaintenanceLimitSpec{
					Policies: []repairmanv1.MaintenancePolicy{repairmanv1.Unschedulable},
				},
			},
			false,
		},
		{
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						corev1.NodeCondition{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			&repairmanv1.MaintenanceLimit{
				Spec: repairmanv1.MaintenanceLimitSpec{
					Policies: []repairmanv1.MaintenancePolicy{repairmanv1.NotReady},
				},
			},
			false,
		},
		{
			corev1.Node{
				Spec: corev1.NodeSpec{
					Unschedulable: true,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						corev1.NodeCondition{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			&repairmanv1.MaintenanceLimit{
				Spec: repairmanv1.MaintenanceLimitSpec{
					Policies: []repairmanv1.MaintenancePolicy{
						repairmanv1.NotReady,
						repairmanv1.Unschedulable,
					},
				},
			},
			false,
		},
	}
	assert, reconciler := testSetup(t)
	for _, tt := range tests {
		assert.Equal(tt.expected, reconciler.nodeAvailable(tt.node, tt.maintenanceLimit))
	}
}
