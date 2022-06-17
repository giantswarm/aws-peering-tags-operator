package ec2

import (
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/giantswarm/aws-peering-tags-operator/pkg/aws/scope"
)

// Service holds a collection of interfaces.
type Service struct {
	scope  scope.VPCScope
	Client ec2iface.EC2API
}

// NewService returns a new service given the NetworkManager api client.
func NewService(clusterScope scope.VPCScope) *Service {
	return &Service{
		scope:  clusterScope,
		Client: scope.NewEC2Client(clusterScope, clusterScope.ARN(), clusterScope.Cluster()),
	}
}
