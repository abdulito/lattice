package componentbuild

import (
	"os"

	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	follow bool
	asJSON bool

	namespaceString string
	url             string
	namespace       types.LatticeNamespace
	userClient      user.Client
	namespaceClient cli.NamespaceClient
)

var Cmd = &cobra.Command{
	Use:  "component-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list component builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		namespaceClient.ComponentBuilds().List()
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get component build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		namespaceClient.ComponentBuilds().Show(id)
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "get component build logs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		namespaceClient.ComponentBuilds().GetLogs(id, follow)
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&namespaceString, "namespace", string(constants.UserSystemNamespace), "namespace to use")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "whether or not to follow the logs")
	getCmd.Flags().BoolVarP(&asJSON, "json", "", false, "whether or not to display output as JSON")
	listCmd.Flags().BoolVarP(&asJSON, "json", "", false, "whether or not to display output as JSON")
}

func initCmd() {
	namespace = types.LatticeNamespace(namespaceString)

	userClient = rest.NewClient(url)
	restClient := userClient.Namespace(namespace)
	namespaceClient = cli.NewNamespaceClient(restClient, asJSON)
}
