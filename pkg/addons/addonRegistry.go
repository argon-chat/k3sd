package addons

import (
	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/argon-chat/k3sd/pkg/utils"
)

type AddonMigrationStatus int

const (
	ToBeApplied AddonMigrationStatus = iota
	Applied
	Changed
	Deleted
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
