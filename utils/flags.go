// Package utils provides utility functions and global variables for flag parsing and configuration.
package utils

import (
	"flag"
	"fmt"
)

// Flags contains parsed boolean flags for feature toggles.
var (
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
		{"cert-manager", false, "Apply the cert-manager YAMLs", "cert-manager"},
		{"traefik", false, "Apply the Traefik YAML", "traefik-values"},
		{"cluster-issuer", false, "Apply the Cluster Issuer YAML, need to specify `domain` in your config json", "clusterissuer"},
		{"gitea", false, "Apply the Gitea YAML", "gitea"},
		{"gitea-ingress", false, "Apply the Gitea Ingress YAML, need to specify `domain` in your config json", "gitea-ingress"},
		{"prometheus", false, "Apply the Prometheus YAML", "prometheus"},
		{"linkerd", false, "Install linkerd", "linkerd"},
		{"linkerd-mc", false, "Install linkerd multicluster(will install linkerd first)", "linkerd-mc"},
	}

	flagPtrs := make(map[string]*bool)
	for _, def := range boolFlags {
		flagPtrs[def.MapKey] = flag.Bool(def.Name, def.Default, def.Description)
	}

	configPath := flag.String("config-path", "", "Path to clusters.json")
	uninstallFlag := flag.Bool("uninstall", false, "Uninstall the cluster")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	verbose := flag.Bool("v", false, "Enable verbose stdout logging")
	helmAtomic := flag.Bool("helm-atomic", false, "Enable --atomic for all Helm operations (rollback on failure)")

	flag.Parse()

	VersionFlag = *versionFlag
	Uninstall = *uninstallFlag
	Verbose = *verbose
	HelmAtomic = *helmAtomic

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
