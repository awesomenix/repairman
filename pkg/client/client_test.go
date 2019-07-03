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

	"testing"

	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	"github.com/awesomenix/repairman/pkg/client"

	//"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	//"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApproveMaintenanceRequest(t *testing.T) {
	assert := assert.New(t)
	repairmanv1.AddToScheme(scheme.Scheme)
	f := fake.NewFakeClient()

	client := &client.Client{
		Client: f,
		Name:   "fakeName",
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

	// check error states
	_, err := client.IsMaintenanceApproved(context.TODO(), "")
	assert.NotNil(err)
	_, err = client.IsMaintenanceApproved(context.TODO(), "dummynode")
	assert.NotNil(err)

	// create maintenance request
	state, err := client.RequestMaintenance(context.TODO(), "dummyNode")
	assert.Nil(err)
	assert.Equal(repairmanv1.Pending, state)

	// set request to approved
	err = client.UpdateMaintenanceState(context.TODO(), "dummyNode", repairmanv1.Approved)
	assert.Nil(err)

	// confirm it is approved and ensure new request isn't generated
	state, err = client.RequestMaintenance(context.TODO(), "dummyNode")
	assert.Nil(err)
	assert.Equal(repairmanv1.Approved, state)

	// confirm it is approved
	isApproved, err := client.IsMaintenanceApproved(context.TODO(), "dummyNode")
	assert.Nil(err)
	assert.Equal(isApproved, true)

	// set request to in-progress and ensure new request cannot be generated against the same node
	err = client.UpdateMaintenanceState(context.TODO(), "dummyNode", repairmanv1.InProgress)
	assert.Nil(err)
	state, err = client.RequestMaintenance(context.TODO(), "dummyNode")
	assert.Nil(err)
	assert.Equal(repairmanv1.InProgress, state)

	//TODO(ganesha): enable tests below by setting namespace to auto generate in tests

	// // set request to completed and ensure new request can be generated against the same node
	// err = client.UpdateMaintenanceState(context.TODO(), "dummyNode", repairmanv1.Completed)
	// assert.Nil(err)
	// state, err = client.RequestMaintenance(context.TODO(), "dummyNode")
	// assert.Nil(err)
	// assert.Equal(repairmanv1.Pending, state)

	// // multiple requests should be able to co-exist
	// state, err = client.RequestMaintenance(context.TODO(), "dummyNode1")
	// assert.Nil(err)
	// assert.Equal(repairmanv1.Pending, state)
	// state, err = client.RequestMaintenance(context.TODO(), "dummyNode2")
	// assert.Nil(err)
	// assert.Equal(repairmanv1.Pending, state)
}
