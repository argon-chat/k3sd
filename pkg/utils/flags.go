package utils

import (
	"flag"
	"fmt"
)

const (
	FlagCertManager   = "cert-manager"
	FlagTraefikValues = "traefik"
	FlagClusterIssuer = "cluster-issuer"
	FlagGitea         = "gitea"
	FlagGiteaIngress  = "gitea-ingress"
	FlagPrometheus    = "prometheus"
	FlagLinkerd       = "linkerd"
	FlagLinkerdMC     = "linkerd-mc"
)

var (
	// Flags contains parsed boolean flags for feature toggles.
	Flags map[string]bool
	// ConfigPath is the path to the cluster config file.
	ConfigPath string
	// Uninstall indicates whether to uninstall the cluster.
	Uninstall bool
	// VersionFlag indicates whether to print version and exit.
	VersionFlag bool
	// Verbose enables verbose logging.
	Verbose bool
	// HelmAtomic enables atomic Helm operations.
	HelmAtomic bool
	// YamlsPath is the prefix path to all YAMLs used for installing additional components.
	// If not set, the program will look for a ./yamls directory or ~/.k3sd/yamls.
	YamlsPath string
)

// boolFlagDef defines a boolean flag for command-line parsing.
type boolFlagDef struct {
	Name        string // The flag name (e.g. "cert-manager")
	Default     bool   // The default value
	Description string // Help text for the flag
	MapKey      string // Key used in the Flags map
}

// ParseFlags parses command-line flags and populates global variables for configuration and feature toggles.
//
// Sets:
//   - Flags: map of feature flags
//   - ConfigPath: path to cluster config file
//   - Uninstall: uninstall mode
//   - VersionFlag: print version and exit
//   - Verbose: enable verbose logging
//   - HelmAtomic: enable atomic Helm operations
func ParseFlags() {
	boolFlags := []boolFlagDef{
		{FlagCertManager, false, "Apply the cert-manager YAMLs", FlagCertManager},
		{FlagTraefikValues, false, "Apply the Traefik YAML", FlagTraefikValues},
		{FlagClusterIssuer, false, "Apply the Cluster Issuer YAML, need to specify `domain` in your config json", FlagClusterIssuer},
		{FlagGitea, false, "Apply the Gitea YAML", FlagGitea},
		{FlagGiteaIngress, false, "Apply the Gitea Ingress YAML, need to specify `domain` in your config json", FlagGiteaIngress},
		{FlagPrometheus, false, "Apply the Prometheus YAML", FlagPrometheus},
		{FlagLinkerd, false, "Install linkerd", FlagLinkerd},
		{FlagLinkerdMC, false, "Install linkerd multicluster(will install linkerd first)", FlagLinkerdMC},
	}

	flagPtrs := make(map[string]*bool)
	for _, def := range boolFlags {
		flagPtrs[def.MapKey] = flag.Bool(def.Name, def.Default, def.Description)
	}

	configPath := flag.String("config-path", "", "Path to clusters.json")
	yamlsPath := flag.String("yamls-path", "", "Prefix path to all YAMLs for installing additional components. If not set, defaults to ./yamls or ~/.k3sd/yamls.")
	uninstallFlag := flag.Bool("uninstall", false, "Uninstall the cluster")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	verbose := flag.Bool("v", false, "Enable verbose stdout logging")
	helmAtomic := flag.Bool("helm-atomic", false, "Enable --atomic for all Helm operations (rollback on failure)")

	flag.Parse()

	VersionFlag = *versionFlag
	Uninstall = *uninstallFlag
	Verbose = *verbose
	HelmAtomic = *helmAtomic
	YamlsPath = *yamlsPath

	Flags = make(map[string]bool)
	for k, ptr := range flagPtrs {
		Flags[k] = *ptr
	}

	if *configPath != "" {
		ConfigPath = *configPath
	} else if !VersionFlag {
		fmt.Println("Must specify --config-path")
		flag.Usage()
	}
}
