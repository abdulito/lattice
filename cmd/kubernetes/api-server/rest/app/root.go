package app

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/api/server/backend"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	kubeconfig       string
	latticeID        string
	port             int
	workingDirectory string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "api-server",
	Run: func(cmd *cobra.Command, args []string) {
		latticeID := v1.LatticeID(latticeID)

		kubernetesBackend, err := backend.NewKubernetesBackend(
			latticeID,
			kubeconfig,
		)
		if err != nil {
			panic(err)
		}

		rest.RunNewRestServer(kubernetesBackend, int32(port), workingDirectory)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.OnInitialize(initCmd)

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initCmd)

	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	RootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	RootCmd.Flags().StringVar(&latticeID, "lattice-id", "", "id of the lattice")
	RootCmd.Flags().StringVar(&workingDirectory, "workingDirectory", "/tmp/lattice-manager-api", "working directory to use")
	RootCmd.Flags().IntVar(&port, "port", 8080, "port to bind to")
}

func initCmd() {
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	goflag.CommandLine.Parse([]string{})
}
