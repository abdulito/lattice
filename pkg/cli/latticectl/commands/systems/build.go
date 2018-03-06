package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
)

type BuildCommand struct {
}

func (c *BuildCommand) Base() (*latticectl.BaseCommand, error) {
	var version string
	cmd := &latticectl.SystemCommand{
		Name: "build",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			BuildSystem(ctx.Client().Systems().SystemBuilds(ctx.SystemID()), version)
		},
	}

	return cmd.Base()
}

func BuildSystem(client client.SystemBuildClient, version string) {
	buildID, err := client.Create(version)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", buildID)
}
