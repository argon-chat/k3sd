package utils

import (
	"flag"
	"fmt"
)

var (
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
	// GenerateFlag indicates whether to launch the interactive TUI to generate a cluster config.
	GenerateFlag bool
)

// boolFlagDef defines a boolean flag for command-line parsing.

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
	configPath := flag.String("config-path", "", "Path to clusters.json")
	yamlsPath := flag.String("yamls-path", "", "Prefix path to all YAMLs for installing additional components. If not set, defaults to ./yamls or ~/.k3sd/yamls.")
	uninstallFlag := flag.Bool("uninstall", false, "Uninstall the cluster")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	verbose := flag.Bool("v", false, "Enable verbose stdout logging")
	helmAtomic := flag.Bool("helm-atomic", false, "Enable --atomic for all Helm operations (rollback on failure)")
	generateFlag := flag.Bool("generate", false, "Launch interactive TUI to generate a cluster config")

	flag.Parse()

	VersionFlag = *versionFlag
	Uninstall = *uninstallFlag
	Verbose = *verbose
	HelmAtomic = *helmAtomic
	YamlsPath = *yamlsPath
	GenerateFlag = *generateFlag

	if *configPath != "" {
		ConfigPath = *configPath
	} else if !VersionFlag {
		fmt.Println("Must specify --config-path")
		flag.Usage()
	}
}
