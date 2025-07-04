package addons

import (
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

// AddonMigration defines the migration logic (Up/Down) for an addon.
//
// Fields:
//   - Up: Function to apply (install/upgrade) the addon.
//   - Down: Function to delete (uninstall) the addon.
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
		Down: DeleteTraefikAddon,
	},
	"prometheus": {
		Up:   ApplyPrometheusAddon,
		Down: DeletePrometheusAddon,
	},
	"cluster-issuer": {
		Up:   ApplyClusterIssuerAddon,
		Down: DeleteClusterIssuerAddon,
	},
	"gitea": {
		Up:   ApplyGiteaAddon,
		Down: DeleteGiteaAddon,
	},
	"linkerd": {
		Up:   ApplyLinkerdAddon,
		Down: DeleteLinkerdAddon,
	},
}
