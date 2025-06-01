# K3SD - K3s Cluster Deployment Tool

K3SD is a command-line tool for creating, managing, and uninstalling K3s Kubernetes clusters across multiple machines.
It automates the deployment of K3s clusters with optional components like cert-manager, Traefik, Prometheus, Gitea, and
Linkerd.

## Features

- Deploy K3s clusters with multiple worker nodes via SSH
- Cross-platform support: Linux (x86_64/arm64), macOS (Apple Silicon), and Windows (x86_64)
- Install and configure optional components:
  - cert-manager
  - Traefik (HTTP/3 enabled, with custom values)
  - Prometheus stack (via Helm)
  - Gitea (with PostgreSQL support and ingress)
  - Linkerd (including multi-cluster, with automated certificate management)
- Generate and manage kubeconfig files for each node
- Clean uninstall of clusters
- Display version information with `--version`
- Verbose logging and atomic Helm operations supported

## Prerequisites

- `kubectl` - [Kubernetes CLI](https://kubernetes.io/docs/tasks/tools/)
- `linkerd` - [Linkerd CLI](https://linkerd.io/2.18/getting-started/#step-1-install-the-cli) (required for Linkerd
  installations)
- `step` - [Certificate management tool](https://smallstep.com/docs/step-cli/installation/) (required for Linkerd)
- `ssh` - SSH client for remote server access

## Installation

Download the appropriate binary for your platform from the [Releases](https://github.com/argon-chat/k3sd/releases) page.

```bash
# Example for Linux x86_64
curl -LO https://github.com/argon-chat/k3sd/releases/latest/download/k3sd-linux-amd64.tar.gz
tar -xzf k3sd-linux-amd64.tar.gz
chmod +x k3sd
sudo mv k3sd /usr/local/bin/
```

## Configuration

Create a JSON configuration file for your clusters. Example:

```jsonc
[
  {
    "address": "192.168.1.10",           // (string) IP address or hostname of the master node
    "user": "root",                      // (string) SSH username for the master node
    "password": "password",              // (string) SSH password for the master node
    "nodeName": "master-1",              // (string) Kubernetes node name for the master
    "labels": {                           // (object) Key-value labels for the master node
      "node-role.kubernetes.io/control-plane": "true" // (string) Example label for control-plane
    },
    "domain": "example.com",             // (string) Domain name for cluster-issuer and Gitea ingress (required if using those features)
    "gitea": {                            // (object) Gitea configuration (only needed if --gitea is used)
      "pg": {                             // (object) PostgreSQL configuration for Gitea
        "user": "gitea",                 // (string) PostgreSQL username for Gitea
        "password": "gitea_password",     // (string) PostgreSQL password for Gitea
        "db": "gitea_db"                  // (string) PostgreSQL database name for Gitea
      }
    },
    "privateNet": false,                  // (boolean) If true, worker nodes are installed by connecting from the master node (using the master's network),
                                           // rather than connecting to each worker directly from your local machine. Set to true if your workers are only
                                           // reachable from the master node (e.g., in a private subnet or behind NAT). Set to false if all nodes are directly
                                           // accessible via SSH from your local machine. This flag determines the installation method for worker nodes.
    "workers": [                          // (array) List of worker node objects
      {
        "address": "192.168.1.11",       // (string) IP address or hostname of the worker node
        "user": "root",                  // (string) SSH username for the worker node
        "password": "password",          // (string) SSH password for the worker node
        "nodeName": "worker-1",          // (string) Kubernetes node name for the worker
        "labels": {                       // (object) Key-value labels for the worker node
          "node-role.kubernetes.io/worker": "true" // (string) Example label for worker
        },
        "done": false                     // (boolean) Internal flag, should be false for new nodes
      }
    ]
  }
]
```

## Usage

### Display Version

```bash
k3sd --version
```

### Create a Cluster

```bash
k3sd --config-path=/path/to/clusters.json
```

### Create a Cluster with Additional Components

```bash
k3sd --config-path=/path/to/clusters.json \
  --cert-manager \
  --traefik \
  --cluster-issuer \
  --prometheus \
  --gitea \
  --gitea-ingress \
  --linkerd \
  --linkerd-mc
```

### Install Linkerd

```bash
k3sd --config-path=/path/to/clusters.json --linkerd
```

### Install Linkerd with Multi-cluster Support

```bash
k3sd --config-path=/path/to/clusters.json --linkerd-mc
```

### Uninstall a Cluster

```bash
k3sd --config-path=/path/to/clusters.json --uninstall
```

## Command-line Options

| Option             | Description                                           |
|--------------------|-------------------------------------------------------|
| `--config-path`    | Path to clusters.json (required)                      |
| `--yamls-path`      | Prefix path to all YAMLs for installing additional components. If not set, the program will look for a ./yamls directory or ~/.k3sd/yamls. |
| `--cert-manager`   | Install cert-manager                                  |
| `--traefik`        | Install Traefik with custom values                    |
| `--cluster-issuer` | Apply ClusterIssuer YAML (requires domain in config)  |
| `--gitea`          | Install Gitea (requires PostgreSQL configuration)     |
| `--gitea-ingress`  | Apply Gitea Ingress (requires domain in config)       |
| `--prometheus`     | Install Prometheus stack (via Helm)                   |
| `--linkerd`        | Install Linkerd with automated certs                  |
| `--linkerd-mc`     | Install Linkerd with multi-cluster support            |
| `--uninstall`      | Uninstall the cluster                                 |
| `--version`        | Print the version and exit                            |
| `-v`               | Enable verbose logging                                |
| `--helm-atomic`    | Enable atomic Helm operations (rollback on failure)   |

## Build from Source

```bash
git clone https://github.com/argon-chat/k3sd.git
cd k3sd
go build -ldflags "-s -w -X 'github.com/argon-chat/k3sd/utils.Version=<version>'" -o k3sd ./cli/main.go
```

For smaller binaries, the build process now strips debug symbols by default. See the CI workflow for details.

## Project Roadmap & Future Milestones


The following table lists planned and completed features/milestones for the project. Status is updated as work progresses.


| Feature / Milestone                                      | Status |
|----------------------------------------------------------|--------|
| Deploy K3s clusters with multiple worker nodes via SSH    | ✅     |
| Cross-platform support (Linux, macOS, Windows)           | ✅     |
| Install cert-manager                                     | ✅     |
| Install Traefik (with custom values, HTTP/3)             | ✅     |
| Install Prometheus stack (via Helm)                      | ✅     |
| Install Gitea (with PostgreSQL and ingress)              | ✅     |
| Install Linkerd (including multi-cluster, auto certs)    | ✅     |
| Generate/manage kubeconfig files for each node           | ✅     |
| Clean uninstall of clusters                              | ✅     |
| Display version information                              | ✅     |
| Verbose logging and atomic Helm operations               | ✅     |
| Support for choosing CNI of choice                       | 🚧     |
| Addon configuration with JSON instead of CLI flags       | 🚧     |
| Add support for more service meshes (e.g., Istio)        | 🚧     |
| Remember/apply config JSON diffs for future changes      | 🚧     |

*Legend: 🚧 = in progress or planned, ✅ = implemented*

## Contributing


Contributions are welcome! Please feel free to submit a Pull Request.

## Adding a New Addon

K3SD uses a unified, extensible addon system. All addons are registered in the `pkg/addons/addonRegistry.go` file and must have the following function signature:

```go
func MyAddon(cluster *types.Cluster, logger *utils.Logger)
```

### Example: Adding Gitea

1. **Create the Addon Function**  
   In `pkg/addons/gitea.go`:
   ```go
   func ApplyGiteaAddon(cluster *types.Cluster, logger *utils.Logger) {
       if !utils.Flags[utils.FlagGitea] {
           return
       }
       kubeconfig := clusterutils.KubeConfigPath(cluster, logger)

       // Example 1: Apply a YAML manifest (with substitutions)
       substitutions := clusterutils.BuildSubstitutions(
           "${POSTGRES_USER}", cluster.Gitea.Pg.Username,
           "${POSTGRES_PASSWORD}", cluster.Gitea.Pg.Password,
           "${POSTGRES_DB}", cluster.Gitea.Pg.DbName,
       )
       clusterutils.ApplyComponentYAML(
           "gitea",                                 // name for logging
           kubeconfig,                               // kubeconfig path
           clusterutils.ResolveYamlPath("gitea.yaml"), // manifest path
           logger,                                   // logger
           substitutions,                            // map[string]string for variable substitution
       )

       // Example 2: Install a Helm chart
       clusterutils.InstallHelmChart(
           kubeconfig,                               // kubeconfig path
           "gitea",                                 // release name
           "default",                               // namespace
           "gitea-charts",                          // repo name
           "https://dl.gitea.io/charts/",           // repo URL
           "gitea",                                 // chart name
           "10.0.0",                                // chart version
           "",                                      // values file path or leave empty
           logger,                                   // logger
       )

       // Wait for deployment to be ready
       clusterutils.WaitForDeploymentReady(kubeconfig, "gitea", "default", logger)
   }
   ```

2. **Register the Addon**  
   In `pkg/addons/addonRegistry.go`:
   ```go
   var AddonRegistry = []AddonFunc{
       // ... other addons
       ApplyGiteaAddon,
   }
   ```

### Guidelines

- Use the provided helpers in `pkg/clusterutils` for manifest application, Helm, and readiness checks.
- Addons should be idempotent and log all actions.
- Addons are invoked automatically if their corresponding CLI flag is set.