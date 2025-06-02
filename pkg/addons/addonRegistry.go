package addons

import (
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// AddonFunc is the function signature for an addon application function.
//
// Parameters:
//
//	cluster: The cluster to apply the addon to.
//	logger: Logger for output.
type AddonFunc func(*types.Cluster, *utils.Logger)

// AddonRegistry is the list of all registered addon application functions.
//
// Each function in this list applies a specific built-in addon to a cluster.
var AddonRegistry = []AddonFunc{
	ApplyCertManagerAddon,
	ApplyTraefikAddon,
	ApplyClusterIssuerAddon,
	ApplyGiteaAddon,
	ApplyPrometheusAddon,
	ApplyLinkerdAddon,
}
