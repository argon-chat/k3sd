package addons

import (
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

type AddonMigration struct {
	Up   func(*types.Cluster, *utils.Logger)
	Down func(*types.Cluster, *utils.Logger)
}

// AddonRegistry maps addon names to their migration logic (Up/Down).
var AddonRegistry = map[string]AddonMigration{
	"cert-manager": {
		Up:   ApplyCertManagerAddon,
		Down: DeleteCertManagerAddon,
	},
	"traefik": {
		Up:   ApplyTraefikAddon,
		Down: DeleteTraefikAddon, // to be implemented
	},
	"prometheus": {
		Up:   ApplyPrometheusAddon,
		Down: DeletePrometheusAddon, // to be implemented
	},
	"cluster-issuer": {
		Up:   ApplyClusterIssuerAddon,
		Down: DeleteClusterIssuerAddon, // to be implemented
	},
	"gitea": {
		Up:   ApplyGiteaAddon,
		Down: DeleteGiteaAddon, // to be implemented
	},
	"linkerd": {
		Up:   ApplyLinkerdAddon,
		Down: DeleteLinkerdAddon, // to be implemented
	},
}
