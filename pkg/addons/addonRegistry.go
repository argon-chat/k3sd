package addons

import (
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// AddonFunc defines the signature for addon application functions.
// Each addon function should accept a pointer to a Cluster and a Logger.
type AddonFunc func(*types.Cluster, *utils.Logger)

// AddonRegistry is the list of all registered addon application functions.
// Add new addons here to make them available for cluster installation.
var AddonRegistry = []AddonFunc{
	ApplyCertManagerAddon,
	ApplyTraefikAddon,
	ApplyClusterIssuerAddon,
	ApplyGiteaAddon,
	ApplyPrometheusAddon,
	ApplyLinkerdAddon,
}
