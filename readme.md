# K3SD - K3s Cluster Deployment Tool

K3SD is a modern, config-driven tool for creating, managing, and uninstalling K3s Kubernetes clusters across multiple machines. It supports fully modular, extensible cluster and addon configuration, and includes a TUI (text user interface) for easy config generation.

---

## Table of Contents

1. [Features](#features)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [TUI Config Generator](#tui-config-generator)
6. [Usage](#usage)
7. [Command-line Options](#command-line-options)
8. [Addon System](#addon-system)
9. [Addon Migration and Idempotency](#addon-migration-and-idempotency)
10. [Linkerd Multicluster Linking](#linkerd-multicluster-linking)
11. [Database and Versioning](#database-and-versioning)
12. [Architecture](#architecture)
13. [Project Roadmap](#project-roadmap)
14. [Contributing](#contributing)
15. [Extending the TUI: Adding New Forms and Inputs](#extending-the-tui-adding-new-forms-and-inputs)

---

## Features

- Deploy K3s clusters with multiple worker nodes via SSH
- Cross-platform: Linux, macOS, Windows
- Fully config-driven: all cluster and addon options are set in a JSON config file
- Built-in addons: cert-manager, Traefik, Prometheus, Gitea, Linkerd, ClusterIssuer, and more
- Custom addon support: install any Helm chart or manifest via config
- TUI (text UI) for interactive config generation
- Clean uninstall of clusters
- Per-node kubeconfig management
- Verbose logging and atomic Helm operations

## Prerequisites

- `kubectl` - [Kubernetes CLI](https://kubernetes.io/docs/tasks/tools/)
- `linkerd` - [Linkerd CLI](https://linkerd.io/2.18/getting-started/#step-1-install-the-cli) (for Linkerd addon)
- `step` - [step CLI](https://smallstep.com/docs/step-cli/installation/) (for Linkerd certs)
- `ssh` - SSH client for remote server access

## Installation

Download the latest binary for your platform from the [Releases](https://github.com/argon-chat/k3sd/releases) page.

```bash
# Example for Linux x86_64
curl -LO https://github.com/argon-chat/k3sd/releases/latest/download/k3sd-linux-amd64.tar.gz
tar -xzf k3sd-linux-amd64.tar.gz
chmod +x k3sd
sudo mv k3sd /usr/local/bin/
```

## Configuration

K3SD uses a single JSON file to describe all clusters and addons. Example:

```json
[
  {
    "address": "10.144.103.55",
    "user": "ubuntu",
    "password": "password123",
    "nodeName": "master",
    "labels": {
      "label1": "value1"
    },
    "domain": "example.com",
    "privateNet": false,
    "workers": [
      {
        "address": "10.144.103.64",
        "user": "ubuntu",
        "password": "password123",
        "nodeName": "worker1",
        "labels": {},
        "done": false
      }
    ],
    "addons": {
      "gitea": {
        "enabled": true,
        "subs": {
          "${POSTGRES_USER}": "gitea",
          "${POSTGRES_PASSWORD}": "changeme",
          "${POSTGRES_DB}": "giteadb"
        }
      },
      "gitea-ingress": {
        "enabled": true,
        "subs": { "${DOMAIN}": "example.com" }
      },
      "cert-manager": { "enabled": true },
      "traefik": { "enabled": false },
      "prometheus": { "enabled": false },
      "cluster-issuer": { "enabled": false },
      "linkerd": { "enabled": false },
      "linkerd-mc": { "enabled": false }
    },
    "customAddons": {
      "somePod": {
        "enabled": false,
        "helm": {
          "chart": "mychart",
          "repo": {
            "name": "myrepo",
            "url": "https://charts.example.com"
          },
          "version": "1.2.3",
          "valuesFile": "./yamls/somepod-values.yaml",
          "namespace": "default"
        },
        "manifest": {
          "path": "./yamls/somepod.yaml",
          "subs": { "KEY": "value" }
        }
      }
    }
  }
]
```

## TUI Config Generator

K3SD includes a built-in TUI for interactively generating cluster configs. Run:

```bash
k3sd -generate
```

This will launch a form-based UI to enter master node info, select addons, and (if needed) configure addon variables. The resulting config is saved as a JSON file.

## Usage

### Display Version

```bash
k3sd --version
```

### Create a Cluster

```bash
k3sd --config-path=/path/to/clusters.json
```

### Uninstall a Cluster

```bash
k3sd --config-path=/path/to/clusters.json --uninstall
```

## Command-line Options

| Option             | Description                                           |
|--------------------|-------------------------------------------------------|
| `--config-path`    | Path to clusters.json (required)                      |
| `--yamls-path`     | Path prefix for YAMLs (default: ./yamls or ~/.k3sd/yamls) |
| `--uninstall`      | Uninstall the cluster                                 |
| `--version`        | Print the version and exit                            |
| `-v`               | Enable verbose logging                                |
| `--helm-atomic`    | Enable atomic Helm operations (rollback on failure)   |
| `-generate`        | Launch the TUI config generator                       |

All addon/component selection is now done via the config file, not CLI flags.

## Addon System

K3SD supports two types of addons:
- **Built-in addons**: Managed by the migration registry with dedicated Up/Down logic. Examples: `cert-manager`, `traefik`, `prometheus`, `gitea`, `cluster-issuer`, `linkerd`, `linkerd-mc`.
- **Custom addons**: User-defined Helm charts or manifests, managed via the `customAddons` map in your config. These use the same migration logic as built-ins.

> **Note:** `gitea-ingress` is not a standalone built-in addon. It is managed as part of the Gitea addon logic and is typically configured as a manifest/ingress resource.

All addons are enabled/disabled via your config file. The migration logic ensures only necessary actions are taken when the cluster config changes, and all install/uninstall flows are robust and idempotent.

## Addon Migration and Idempotency

K3SD uses an enum-based migration status:
- `AddonApply`: Install or upgrade the addon.
- `AddonDelete`: Uninstall the addon.
- `AddonNoop`: No action needed (state unchanged).

The function `ComputeAddonMigrationStatus` determines the correct action for each addon based on the current and previous cluster state. This ensures safe, repeatable operations and supports upgrades, rollbacks, and config diffs.

### Adding a New Built-in Addon
1. Implement `Up` and `Down` functions in `pkg/addons/youraddon.go`.
2. Register your addon in `pkg/addons/addonRegistry.go`.
3. Add config keys and substitutions as needed.

### Adding a Custom Addon
Add a new entry to the `customAddons` map in your config file, specifying either a Helm chart or manifest (or both). No code changes are required for most custom addons. Both Helm and manifest custom addons are supported and can be enabled/disabled independently.

---

## Linkerd Multicluster Linking

K3SD now supports robust, idempotent Linkerd multicluster linking. If you enable both `linkerd` and `linkerd-mc` addons and specify a `linksTo` array in your cluster config, K3SD will automatically link your clusters using the correct kubeconfigs and Linkerd CLI commands.

- The system checks for existing links and unlinks as needed.
- Linking/unlinking is idempotent and robust against config changes.
- Handles error cases and uninstall order correctly.

**Example cluster config:**

```json
{
  "context": "cluster-1",
  "nodeName": "cluster-1-master",
  ...
  "addons": {
    "linkerd": { "enabled": true },
    "linkerd-mc": { "enabled": true }
  },
  "linksTo": ["cluster-2", "cluster-3"]
}
```

> **Note:** The `linksTo` field in your config should contain the context names of the clusters you wish to link, not their IP addresses.

---

## Database and Versioning

K3SD stores cluster state and version history in a local SQLite database (via GORM). This enables:
- Safe upgrades and rollbacks
- Accurate migration logic for addons
- Tracking of cluster changes over time

---

## Architecture

K3SD is organized into several key packages:

- **pkg/cluster**: Handles cluster creation, worker join, uninstall, and main orchestration logic.
- **pkg/addons**: Built-in and custom addon management, including migration logic (Up/Down), registry, and linking (Linkerd).
- **pkg/clusterutils**: Utilities for YAML/Helm apply/delete, SSH, manifest handling, and migration status computation.
- **pkg/types**: All config and runtime types (Cluster, AddonConfig, CustomAddonConfig, etc).
- **pkg/db**: Cluster state/versioning with SQLite (via GORM).
- **pkg/utils**: Logging, CLI flags, version, and helpers.
- **pkg/k8s**: Kubeconfig and Kubernetes-specific helpers.

---

## Project Roadmap

| Feature / Milestone                                      | Status |
|----------------------------------------------------------|--------|
| Deploy K3s clusters with multiple worker nodes via SSH    | ✅     |
| Cross-platform support (Linux, macOS, Windows)           | ✅     |
| Built-in addon system (config-driven)                    | ✅     |
| Custom addon support (Helm/manifest)                     | ✅     |
| TUI config generator                                     | ✅     |
| Clean uninstall of clusters                              | ✅     |
| Per-node kubeconfig management                           | ✅     |
| Verbose logging and atomic Helm operations               | ✅     |
| Support for choosing CNI of choice                       | 🚧     |
| Add support for more service meshes (e.g., Istio)        | 🚧     |
| Remember/apply config JSON diffs for future changes      | 🚧     |

*Legend: 🚧 = in progress or planned, ✅ = implemented*

## Contributing

Contributions are welcome! To get started:

1. Fork the repo and create a feature branch.
2. Make your changes (see below for addon guidelines).
3. Run `go build`, `go vet`, and `golangci-lint run` to ensure code quality.
4. Submit a pull request with a clear description.

### Adding a New Built-in Addon

1. Create your addon logic in `pkg/addons/youraddon.go` as a function:
   ```go
   func ApplyYourAddon(cluster *types.Cluster, logger *utils.Logger) { /* ... */ }
   ```
2. Register it in `pkg/addons/addonRegistry.go`.
3. Add config keys and substitutions as needed (see other addons for examples).

### Adding a Custom Addon (No Code Required)

Add a new entry to the `customAddons` map in your config file, specifying either a Helm chart or manifest (or both). See the config example above.

### Guidelines

- Use helpers in `pkg/clusterutils` for manifest/Helm operations.
- Addons should be idempotent and log all actions.
- Document any new config keys in the README.

---

## Extending the TUI: Adding New Forms and Inputs

The TUI is designed to be modular and easily extensible. To add a new input field or a new form (e.g., for a new addon), follow these steps:

### Adding a New Input Field to the Cluster Form

1. **Edit the `clusterFields` array** in `cli/tui/generate.go`:
   ```go
   var clusterFields = []FieldDef{
       {"Master node IP", "", false},
       {"Master SSH user", "", false},
       // ...
       {"My New Field", "default-value", false}, // <-- Add your field here
   }
   ```
2. **No further code changes are needed.** The field will automatically appear in the TUI and be included in the generated config.

### Adding a New Addon Form (with custom inputs)

1. **Define your addon fields** as a `[]FieldDef`:
   ```go
   var myAddonFields = []FieldDef{
       {"MY_OPTION", "default", false},
       {"MY_SECRET", "", true},
   }
   ```
2. **Create a form function using the generic builder:**
   ```go
   func buildMyAddonForm(app *tview.Application, onBack func(), onDone func(subs map[string]string)) *tview.Form {
       return buildAddonSubsForm(app, "MyAddon Configuration", myAddonFields, onBack, onDone)
   }
   ```
3. **Add your addon to the `addonList` array**:
   ```go
   var addonList = []string{
       "gitea", "myaddon", // ...
   }
   ```
4. **Update the logic in `buildClusterForm`** to call your new form when your addon is selected (see how Gitea is handled for an example).

### Guidelines
- All field definitions are arrays at the top of `generate.go`.
- Use the `FieldDef` struct for each field: `{Label, Default, IsPassword}`.
- Use the `buildAddonSubsForm` helper for any new addon form.
- No need to modify core logic—just add to arrays and call the generic builder.

This approach keeps the code DRY, modular, and easy to maintain. For more advanced flows (multi-step forms, validation, etc.), follow the same pattern: define your fields, use the generic builder, and handle the result in the main flow.

---
