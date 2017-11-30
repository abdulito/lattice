package app

import (
	"fmt"
	"os"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	restclient "github.com/mlab-lattice/system/pkg/manager/api/client/rest"
	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/user"

	"github.com/spf13/cobra"
)

const (
	devDockerRegistry = "gcr.io/lattice-dev"
)

var (
	workingDir      string
	namespaceString string
	url             string
	namespace       coretypes.LatticeNamespace
	userClient      *user.Client
	namespaceClient *user.NamespaceClient
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lattice-system",
	Short: "The lattice-system CLI tool",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initCmd)
	RootCmd.PersistentFlags().StringVar(&workingDir, "working-directory", "/tmp/lattice-system/", "path where subcommands will use as their working directory")
	RootCmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	RootCmd.PersistentFlags().StringVar(&namespaceString, "namespace", string(coreconstants.UserSystemNamespace), "namespace to use")
}

func initCmd() {
	err := os.MkdirAll(workingDir, 0770)
	if err != nil {
		panic(fmt.Errorf("unable to create log-path: %v", err))
	}

	namespace = coretypes.LatticeNamespace(namespaceString)

	userClient = restclient.NewUserClient(url)
	namespaceClient = userClient.Namespace(namespace)
}
