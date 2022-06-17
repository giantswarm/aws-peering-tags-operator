package tagger

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-peering-tags-operator/pkg/aws/scope"
	"github.com/giantswarm/aws-peering-tags-operator/pkg/aws/services/cloudformation"
	"github.com/giantswarm/aws-peering-tags-operator/pkg/aws/services/ec2"
)

type TaggerService struct {
	Client client.Client
	Scope  *scope.ClusterScope

	Cloudformation *cloudformation.Service
	EC2            *ec2.Service
}

func New(scope *scope.ClusterScope, client client.Client) *TaggerService {
	return &TaggerService{
		Scope:  scope,
		Client: client,

		Cloudformation: cloudformation.NewService(scope),
		EC2:            ec2.NewService(scope),
	}
}

func (s *TaggerService) Reconcile(ctx context.Context) error {
	s.Scope.Info("Reconciling AWSCluster CR for tagging VPC peering")

	stacksInput := &cf.DescribeStacksInput{StackName: aws.String(fmt.Sprintf("cluster-%s-tccp", s.Scope.ClusterName()))}
	stacksOutput, err := s.Cloudformation.Client.DescribeStacks(stacksInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cf.ErrCodeStackNotFoundException:
				return nil
			}
		}
		return err
	}
	stackStatus := stacksOutput.Stacks[0].StackStatus
	if *stackStatus != cf.StackStatusCreateComplete {
		return nil
	}

	peeringConnectionID := ""
	for _, o := range stacksOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VPCPeeringConnectionID" {
			peeringConnectionID = *o.OutputValue
		}
	}
	peeringConnectionsOutput, err := s.EC2.Client.DescribeVpcPeeringConnections(&awsec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: []*string{aws.String(peeringConnectionID)},
	})
	if err != nil {
		return err
	}
	if peeringConnectionsOutput.VpcPeeringConnections[0].Tags != nil {
		s.Scope.Logger.Info("Already tagged VPC peering connection, skipping", s.Scope.ClusterNamespace(), s.Scope.ClusterName())
		return nil
	}

	tags := []*awsec2.Tag{}
	for _, t := range stacksOutput.Stacks[0].Tags {
		tags = append(tags, &awsec2.Tag{Key: t.Key, Value: t.Value})
	}

	_, err = s.EC2.Client.CreateTags(&awsec2.CreateTagsInput{
		Resources: []*string{aws.String(peeringConnectionID)},
		Tags:      tags,
	})
	if err != nil {
		return err
	}
	s.Scope.Logger.Info("Tagged VPC peering connection", s.Scope.ClusterNamespace(), s.Scope.ClusterName())

	return nil
}
