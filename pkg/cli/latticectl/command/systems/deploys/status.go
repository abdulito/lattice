package deploys

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

// GetDeploysSupportedFormats is the list of printer.Formats supported
// by the GetDeploy function.
var GetDeploysSupportedFormats = []printer.Format{
	printer.FormatDefault,
	printer.FormatJSON,
	printer.FormatTable,
}

type StatusCommand struct {
}

func (c *StatusCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: GetDeploysSupportedFormats,
	}
	var watch bool

	cmd := &lctlcommand.DeployCommand{
		Name: "status",
		Flags: command.Flags{
			output.Flag(),
			&command.BoolFlag{
				Name:    "watch",
				Short:   "w",
				Default: false,
				Target:  &watch,
			},
		},
		Run: func(ctx lctlcommand.DeployCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			c := ctx.Client().Systems().Rollouts(ctx.SystemID())

			if watch {
				WatchDeploy(c, ctx.DeployID(), format, os.Stdout)
			}

			err = GetDeploy(c, ctx.DeployID(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetDeploy(client client.RolloutClient, deployID types.SystemRolloutID, format printer.Format, writer io.Writer) error {
	deploy, err := client.Get(deployID)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", deploy)
	return nil
}

func WatchDeploy(client client.RolloutClient, deployID types.SystemRolloutID, format printer.Format, writer io.Writer) {
	deploy, err := client.Get(deployID)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy)
}
