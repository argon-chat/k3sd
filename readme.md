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
7. [Addon System](#addon-system)
8. [Custom Addons](#custom-addons)
9. [Uninstalling Clusters](#uninstalling-clusters)
10. [Build from Source](#build-from-source)
11. [Project Roadmap](#project-roadmap)
12. [Contributing](#contributing)
13. [Extending the TUI: Adding New Forms and Inputs](#extending-the-tui-adding-new-forms-and-inputs)
14. [Linkerd Multicluster Linking](#linkerd-multicluster-linking)

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

All built-in addons are enabled/configured via the `addons` map in your cluster config. Supported built-in addons:

- `cert-manager`
- `traefik`
- `prometheus`
- `gitea`
- `gitea-ingress`
- `cluster-issuer`
- `linkerd`
- `linkerd-mc`

Each addon can be enabled/disabled and provided with variable substitutions via the `subs` map. See the config example above.

## Custom Addons

You can install any Helm chart or manifest by adding entries to the `customAddons` map in your config. Example:

```json
"customAddons": {
  "myaddon": {
    "enabled": true,
    "helm": {
      "chart": "mychart",
      "repo": { "name": "myrepo", "url": "https://charts.example.com" },
      "version": "1.2.3",
      "valuesFile": "./yamls/myaddon-values.yaml",
      "namespace": "default"
    },
    "manifest": {
      "path": "./yamls/myaddon.yaml",
      "subs": { "KEY": "value" }
    }
  }
}
```

## Uninstalling Clusters

To uninstall all clusters defined in your config:

```bash
k3sd --config-path=/path/to/clusters.json --uninstall
```

## Build from Source

```bash
git clone https://github.com/argon-chat/k3sd.git
cd k3sd
go build -ldflags "-s -w -X 'github.com/argon-chat/k3sd/utils.Version=<version>'" -o k3sd ./cli/main.go
```

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

## Linkerd Multicluster Linking

K3SD now supports automated Linkerd multicluster linking. If you enable both `linkerd` and `linkerd-mc` addons and specify a `linksTo` array in your cluster config, K3SD will automatically link your clusters using the correct kubeconfigs and Linkerd CLI commands.

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
  "linksTo": ["cluster-2-ip-address", "cluster-3-ip-address"]
}
```

**How it works:**
- For each entry in `linksTo`, K3SD runs:
  ```
  linkerd multicluster link --set "enableHeadlessServices=true" --log-level="debug" --cluster-name=<context> --api-server-address=https://<cluster>:6443 | kubectl apply -f -
  ```
- This enables seamless multicluster service mesh federation with Linkerd.

---
