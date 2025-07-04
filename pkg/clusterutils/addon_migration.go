package clusterutils

import (
	"github.com/argon-chat/k3sd/pkg/types"
)

// AddonMigrationStatus represents the migration status of an addon (noop, apply, delete).
type AddonMigrationStatus int

const (
	// AddonNoop indicates no migration action is needed for the addon.
	AddonNoop AddonMigrationStatus = iota
	// AddonApply indicates the addon should be applied (installed or upgraded).
	AddonApply
	// AddonDelete indicates the addon should be deleted (uninstalled).
	AddonDelete
)

func ComputeAddonMigrationStatus(name string, cluster *types.Cluster, oldVersion *types.Cluster, customAddon bool) AddonMigrationStatus {
	if customAddon {
		return customAddonMigrationStatus(name, cluster, oldVersion)
	}
	return registryAddonMigrationStatus(name, cluster, oldVersion)
}

func registryAddonMigrationStatus(name string, cluster *types.Cluster, oldVersion *types.Cluster) AddonMigrationStatus {
	addon, ok := cluster.Addons[name]

	if oldVersion == nil {
		if ok && addon.Enabled {
			return AddonApply
		} else {
			return AddonDelete
		}
	} else {
		oldAddon, oldOk := oldVersion.Addons[name]
		if oldOk && !ok {
			return AddonDelete
		}
		if oldOk && ok && oldAddon.Enabled == addon.Enabled {
			return AddonNoop
		}
		if ok {
			if addon.Enabled {
				return AddonApply
			} else {
				return AddonDelete
			}
		}
	}
	return AddonNoop
}

func customAddonMigrationStatus(name string, cluster *types.Cluster, oldVersion *types.Cluster) AddonMigrationStatus {
	customAddon, ok := cluster.CustomAddons[name]
	if !ok {
		return AddonNoop
	}

	if oldVersion == nil {
		if customAddon.Enabled && (customAddon.Manifest != nil || customAddon.Helm != nil) {
			return AddonApply
		} else {
			return AddonDelete
		}
	}

	oldCustomAddon, oldOk := oldVersion.CustomAddons[name]
	if oldOk && !ok {
		return AddonDelete
	}
	if oldOk && ok && oldCustomAddon.Enabled == customAddon.Enabled {
		return AddonNoop
	}
	if ok {
		if customAddon.Enabled && (customAddon.Manifest != nil || customAddon.Helm != nil) {
			return AddonApply
		} else {
			return AddonDelete
		}
	}
	return AddonNoop
}
