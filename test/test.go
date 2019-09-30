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

package test

import (
	"context"
	"fmt"
	"math/rand"

	repairmanv1 "github.com/awesomenix/repairman/api/v1"
	"github.com/awesomenix/repairman/controllers"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

var names []string

// NewRequest ...
func NewRequest() *repairmanv1.MaintenanceRequest {
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

// ReconcileML ...
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

// ReconcileMR ...
func ReconcileMR(f ctrlclient.Client) {
	mrreconciler := &controllers.MaintenanceRequestReconciler{
		Client:        f,
		EventRecorder: &record.FakeRecorder{},
		Log:           ctrl.Log,
	}

	for _, name := range names {
		mrreconciler.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name: name,
			},
		})
	}
}
