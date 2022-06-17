package scope

import (
	"github.com/giantswarm/aws-peering-tags-operator/pkg/aws"
)

// CloudformationScope is a scope for use with the Cloudformation reconciling service in cluster
type CloudformationScope interface {
	aws.ClusterScoper
}

// VPCScope is a scope for use with the NetworkManager reconciling service in cluster
type VPCScope interface {
	aws.ClusterScoper
}
