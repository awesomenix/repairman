package client

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	scheme   = runtime.NewScheme()
	nodeType = "node"
)

func init() {
	corev1.AddToScheme(scheme)
	repairmanv1.AddToScheme(scheme)
}

// Client implementation for the repairman client
type Client struct {
	Name string
	client.Client
	NewRequest func() *repairmanv1.MaintenanceRequest
}

// New intializes repairman client
func New(clientName string) (*Client, error) {
	if clientName == "" {
		return nil, errors.New("client name required")
	}
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	client, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return &Client{
		Name:       clientName,
		Client:     client,
		NewRequest: newRequest,
	}, nil
}

// IsEnabled checks if repairman is enabled on server
func (c *Client) IsEnabled(rtype string) (bool, error) {
	if rtype != nodeType {
		return false, nil
	}

	mrCustomResourceDef := &apiextensions.CustomResourceDefinition{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: "maintenancerequests.repairman.k8s.io"}, mrCustomResourceDef)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return false, nil
}

// RequestMaintenance returns existing request, else creates a new one
func (c *Client) requestMaintenance(ctx context.Context, name, rtype string) (*repairmanv1.MaintenanceRequest, error) {
	request, err := c.getRequest(ctx, name, rtype)
	if err != nil {
		return nil, err
	}
	if request != nil {
		return request, nil
	}

	request = c.NewRequest()
	request.Namespace = c.Name
	request.Spec = repairmanv1.MaintenanceRequestSpec{
		Name:  name,
		Type:  rtype,
		State: repairmanv1.Pending,
	}

	return request, c.Client.Create(ctx, request)
}

func newRequest() *repairmanv1.MaintenanceRequest {
	return &repairmanv1.MaintenanceRequest{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "mr-",
		},
	}
}

func (c *Client) getRequest(ctx context.Context, name, rtype string) (*repairmanv1.MaintenanceRequest, error) {
	maintenanceRequests := &repairmanv1.MaintenanceRequestList{}
	err := c.Client.List(ctx, maintenanceRequests, client.InNamespace(c.Name))
	if err != nil {
		return nil, err
	}

	for _, request := range maintenanceRequests.Items {
		if strings.EqualFold(request.Spec.Name, name) &&
			strings.EqualFold(request.Spec.Type, rtype) &&
			request.Spec.State != repairmanv1.Completed {
			return &request, nil
		}
	}

	return nil, nil
}

// validate checks if node name is valid
// should also validate against current list of nodes

func (c *Client) validateNode(ctx context.Context, name string) error {
	nodelist := &corev1.NodeList{}
	err := c.List(ctx, nodelist)
	if err != nil {
		return err
	}
	for _, node := range nodelist.Items {
		if strings.EqualFold(node.Name, name) {
			return nil
		}
	}
	return errors.Errorf("invalid node name in %d node list", len(nodelist.Items))
}

func (c *Client) validate(ctx context.Context, name, rtype string) error {
	switch rtype {
	case nodeType:
		return c.validateNode(ctx, name)
	}

	return nil
}

// IsMaintenanceApproved checks if a maintenance request on a name and type is approved
func (c *Client) IsMaintenanceApproved(ctx context.Context, name, rtype string) (bool, error) {
	if err := c.validate(ctx, name, rtype); err != nil {
		return false, err
	}
	request, err := c.requestMaintenance(ctx, name, rtype)
	if err != nil {
		return false, err
	}
	if request.Spec.State == repairmanv1.Approved {
		return true, nil
	}

	return false, nil
}

// UpdateMaintenanceState sets desired state if a live (not completed) maintenance request exists
func (c *Client) UpdateMaintenanceState(ctx context.Context, name, rtype string, desiredState repairmanv1.MaintenanceState) error {
	if desiredState == repairmanv1.Approved {
		return errors.New("cannot set state as approved")
	}
	if err := c.validate(ctx, name, rtype); err != nil {
		return err
	}
	request, err := c.getRequest(ctx, name, rtype)
	if err != nil {
		return err
	}
	if request == nil {
		return errors.New("request must exist before state can be updated")
	}
	switch request.Spec.State {
	case repairmanv1.Pending:
		return errors.New("request is not yet approved")
	case repairmanv1.Completed:
		return errors.New("request is already completed")
	}
	request.Spec.State = desiredState

	return c.Client.Update(ctx, request)
}
