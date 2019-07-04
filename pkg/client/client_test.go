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

package client_test

import (
	"context"
	"fmt"
	"math/rand"

	"testing"

	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	"github.com/awesomenix/repairman/pkg/client"
	"github.com/awesomenix/repairman/pkg/controllers"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

var names []string

func newRequest() *repairmanv1.MaintenanceRequest {
	b := make([]byte, 31)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	names = append(names, string(b))
	return &repairmanv1.MaintenanceRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(b),
		},
	}
}

func ReconcileML(f ctrlclient.Client, assert *assert.Assertions) {
	reconciler := &controllers.MaintenanceLimitReconciler{
		Client:        f,
		EventRecorder: &record.FakeRecorder{},
		Log:           ctrl.Log,
	}

	for i := 0; i < 2; i++ {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("dummynode%d", i),
			},
		}
		err := f.Create(context.TODO(), node)
		assert.Nil(err)
	}

	ml := &repairmanv1.MaintenanceLimit{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node",
		},
		Spec: repairmanv1.MaintenanceLimitSpec{
			Limit: 10,
		},
	}
	err := f.Create(context.TODO(), ml)
	assert.Nil(err)
	err = reconciler.UpdateMaintenanceLimits(context.TODO(), reconciler.Log, ml)
	assert.Nil(err)
}

func ReconcileMR(f ctrlclient.Client) {
	mrreconciler := &controllers.MaintenanceRequestReconciler{
		Client:        f,
		EventRecorder: &record.FakeRecorder{},
		Log:           ctrl.Log,
	}

	for _, name := range names {
		mrreconciler.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "fakeName",
				Name:      name,
			},
		})
	}
}

func TestClient(t *testing.T) {
	assert := assert.New(t)
	repairmanv1.AddToScheme(scheme.Scheme)
	f := fake.NewFakeClient()

	client := &client.Client{
		Client:     f,
		Name:       "fakeName",
		NewRequest: newRequest,
	}

	const nodeType = "node"
	ReconcileML(f, assert)

	isApproved, err := client.IsMaintenanceApproved(context.TODO(), "", nodeType)
	assert.NotNil(err)
	assert.False(isApproved)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode0", nodeType)
	assert.Nil(err)
	assert.False(isApproved)

	ReconcileMR(f)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode0", nodeType)
	assert.Nil(err)
	assert.True(isApproved)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode1", nodeType)
	assert.Nil(err)
	assert.False(isApproved)

	ReconcileMR(f)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode1", nodeType)
	assert.Nil(err)
	assert.False(isApproved)

	err = client.UpdateMaintenanceState(context.TODO(), "dummynode0", nodeType, repairmanv1.Approved)
	assert.NotNil(err)
	err = client.UpdateMaintenanceState(context.TODO(), "dummynode0", nodeType, repairmanv1.InProgress)
	assert.Nil(err)

	err = client.UpdateMaintenanceState(context.TODO(), "dummynode1", nodeType, repairmanv1.InProgress)
	assert.NotNil(err)

	err = client.UpdateMaintenanceState(context.TODO(), "dummynode0", nodeType, repairmanv1.Completed)
	assert.Nil(err)

	ReconcileMR(f)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode1", nodeType)
	assert.Nil(err)
	assert.True(isApproved)
}
