package scope

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	ARN              string
	Cluster          runtime.Object
	ClusterName      string
	ClusterNamespace string
	Region           string

	Logger  logr.Logger
	Session awsclient.ConfigProvider
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	if params.ARN == "" {
		return nil, errors.New("failed to generate new scope from emtpy string ARN")
	}
	if params.Cluster == nil {
		return nil, errors.New("failed to generate new scope from nil Cluster")
	}
	if params.ClusterName == "" {
		return nil, errors.New("failed to generate new scope from emtpy string ClusterName")
	}
	if params.ClusterNamespace == "" {
		return nil, errors.New("failed to generate new scope from emtpy string ClusterNamespace")
	}
	if params.Region == "" {
		return nil, errors.New("failed to generate new scope from emtpy string Region")
	}

	session, err := sessionForRegion(params.Region)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}
	awsClientConfig := &aws.Config{Credentials: stscreds.NewCredentials(session, params.ARN)}

	stsClient := sts.New(session, awsClientConfig)
	o, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sts client")
	}

	params.Logger.Info(fmt.Sprintf("assumed role %s", *o.Arn))

	return &ClusterScope{
		assumeRole:       params.ARN,
		cluster:          params.Cluster,
		clusterName:      params.ClusterName,
		clusterNamespace: params.ClusterNamespace,

		Logger:  params.Logger,
		session: session,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	assumeRole       string
	cluster          runtime.Object
	clusterName      string
	clusterNamespace string
	region           string

	logr.Logger
	session awsclient.ConfigProvider
}

// ARN returns the AWS SDK assumed role.
func (s *ClusterScope) ARN() string {
	return s.assumeRole
}

// Cluster returns the AWS infrastructure cluster object.
func (s *ClusterScope) Cluster() runtime.Object {
	return s.cluster
}

// ClusterName returns the name of AWS infrastructure cluster object.
func (s *ClusterScope) ClusterName() string {
	return s.clusterName
}

// ClusterNameSpace returns the namespace of AWS infrastructure cluster object.
func (s *ClusterScope) ClusterNamespace() string {
	return s.clusterNamespace
}

// Region returns the region of the AWS infrastructure cluster object.
func (s *ClusterScope) Region() string {
	return s.region
}

// Session returns the AWS SDK session.
func (s *ClusterScope) Session() awsclient.ConfigProvider {
	return s.session
}
