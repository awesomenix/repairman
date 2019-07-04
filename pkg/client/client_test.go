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

	"testing"

	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	"github.com/awesomenix/repairman/pkg/client"

	"github.com/awesomenix/repairman/pkg/test"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestClient(t *testing.T) {
	assert := assert.New(t)
	repairmanv1.AddToScheme(scheme.Scheme)
	f := fake.NewFakeClient()

	client := &client.Client{
		Client:     f,
		Name:       "fakeName",
		NewRequest: test.NewRequest,
	}

	const nodeType = "node"
	test.ReconcileML(f, assert)

	isApproved, err := client.IsMaintenanceApproved(context.TODO(), "", nodeType)
	assert.NotNil(err)
	assert.False(isApproved)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode0", nodeType)
	assert.Nil(err)
	assert.False(isApproved)

	test.ReconcileMR(f)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode0", nodeType)
	assert.Nil(err)
	assert.True(isApproved)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode1", nodeType)
	assert.Nil(err)
	assert.False(isApproved)

	test.ReconcileMR(f)

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

	test.ReconcileMR(f)

	isApproved, err = client.IsMaintenanceApproved(context.TODO(), "dummynode1", nodeType)
	assert.Nil(err)
	assert.True(isApproved)
}
