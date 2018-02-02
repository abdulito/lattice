package deprovision

import (
	"fmt"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/cluster/provisioner"

	"github.com/spf13/cobra"
)

var (
	workDir     string
	force       bool
	backend     string
	backendVars []string
)

var Cmd = &cobra.Command{
	Use:   "deprovision [PROVIDER] [NAME]",
	Short: "Deprovision a system",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		name := args[1]

		var provisioner provisioner.Interface
		switch backend {
		case constants.BackendTypeKubernetes:
			var err error
			provisioner, err = getKubernetesProvisioner(provider)
			if err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("unsupported backend %v", backend))
		}

		err := provisioner.Deprovision(name, force)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	Cmd.Flags().StringVar(&workDir, "work-directory", "/tmp/lattice/cluster", "path where subcommands will use as their working directory")
	Cmd.Flags().BoolVar(&force, "force", false, "if set, attempt to deprovision the cluster without tearing down systems")
	Cmd.Flags().StringVar(&backend, "backend", constants.BackendTypeKubernetes, "lattice backend to use")
	Cmd.Flags().StringArrayVar(&backendVars, "backend-var", nil, "additional variables to pass in to the backend")
}
