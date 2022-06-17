package cloudformation

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"

	"github.com/giantswarm/aws-peering-tags-operator/pkg/aws/scope"
)

// Service holds a collection of interfaces.
type Service struct {
	scope  scope.CloudformationScope
	Client cloudformationiface.CloudFormationAPI
}

// NewService returns a new service given the Cloudformation api client.
func NewService(clusterScope scope.CloudformationScope) *Service {
	return &Service{
		scope:  clusterScope,
		Client: scope.NewCloudformationClient(clusterScope, clusterScope.ARN(), clusterScope.Cluster()),
	}
}
