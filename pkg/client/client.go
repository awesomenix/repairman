package client

import (
	"context"
	"errors"
	"strings"

	repairmanv1 "github.com/awesomenix/repairman/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	scheme = runtime.NewScheme()
	Node   = "node"
)

func init() {
	corev1.AddToScheme(scheme)
	repairmanv1.AddToScheme(scheme)
}

// Client implementation for the repairman client
type Client struct {
	Name string
	client.Client
}

// NewClient intializes repairman client
func NewClient(clientName string) (*Client, error) {
	if clientName == "" {
		return nil, errors.New("Client name required")
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
		Name:   clientName,
		Client: client,
	}, nil
}

// RequestMaintenance approves maintenance for a given node or returns the maintenance state if the node is already being updated
func (c *Client) RequestMaintenance(ctx context.Context, nodeName string) (repairmanv1.MaintenanceState, error) {
	if nodeName == "" {
		return "", errors.New("Node name required")
	}

	request, err := c.getLiveRequestByClientName(ctx, nodeName)
	if err != nil {
		return "", err
	}
	if request != nil {
		return request.Spec.State, nil
	}

	request = &repairmanv1.MaintenanceRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    c.Name,
			GenerateName: "mr-",
		},
		Spec: repairmanv1.MaintenanceRequestSpec{
			Name:  nodeName,
			Type:  Node,
			State: repairmanv1.Pending,
		},
	}

	return repairmanv1.Pending, c.Client.Create(ctx, request)
}

func (c *Client) generateRequest(nodeName string) *repairmanv1.MaintenanceRequest {
	return &repairmanv1.MaintenanceRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    c.Name,
			GenerateName: "mr-",
		},
		Spec: repairmanv1.MaintenanceRequestSpec{
			Name:  nodeName,
			Type:  Node,
			State: repairmanv1.Pending,
		},
	}
}

func (c *Client) getLiveRequestByClientName(ctx context.Context, nodeName string) (*repairmanv1.MaintenanceRequest, error) {
	maintenanceRequests := &repairmanv1.MaintenanceRequestList{}
	err := c.Client.List(ctx, maintenanceRequests, client.InNamespace(c.Name))
	if err != nil {
		return nil, err
	}

	for _, request := range maintenanceRequests.Items {
		if strings.EqualFold(request.Spec.Name, nodeName) && request.Spec.State != repairmanv1.Completed {
			return &request, nil
		}
	}

	return nil, nil
}

// IsMaintenanceApproved checks if a maintenance request on a node is approved
func (c *Client) IsMaintenanceApproved(ctx context.Context, nodeName string) (bool, error) {
	if nodeName == "" {
		return false, errors.New("Node name required")
	}
	request, err := c.getLiveRequestByClientName(ctx, nodeName)
	if err != nil {
		return false, err
	}
	if request == nil {
		return false, errors.New("A live request must exist before state can be updated. Please use RequestMaintenance to create a maintenance request")
	}
	if request.Spec.State == repairmanv1.Approved {
		return true, nil
	}

	return false, nil
}

// UpdateMaintenanceState sets desired state if a live (not completed) maintenance request exists
func (c *Client) UpdateMaintenanceState(ctx context.Context, nodeName string, desiredState repairmanv1.MaintenanceState) error {
	if nodeName == "" {
		return errors.New("Node name required")
	}
	request, err := c.getLiveRequestByClientName(ctx, nodeName)
	if err != nil {
		return err
	}
	if request == nil || request.Spec.State == repairmanv1.Completed {
		return errors.New("A live request must exist before state can be updated. Please use RequestMaintenance to create a maintenance request")
	}
	request.Spec.State = desiredState

	return c.Client.Update(ctx, request)
}
