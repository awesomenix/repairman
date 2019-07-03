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

package controllers_test

import (
	"context"
	"fmt"

	"testing"

	repairmanv1 "github.com/awesomenix/repairman/api/v1"
	"github.com/awesomenix/repairman/controllers"

	//"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUpdateMaitenanceLimits(t *testing.T) {
	assert := assert.New(t)
	repairmanv1.AddToScheme(scheme.Scheme)
	f := fake.NewFakeClient()

	reconciler := &controllers.MaintenanceLimitReconciler{
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
	assert.Contains(err.Error(), "unsupported")

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
	assert.Equal(ml.Status.Limit, uint(1))

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
	assert.Equal(ml.Status.Limit, uint(5))
}
