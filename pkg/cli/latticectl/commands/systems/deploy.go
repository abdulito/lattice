package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeployCommand struct {
}

func (c *DeployCommand) Base() (*latticectl.BaseCommand, error) {
	var buildID string
	var version string
	cmd := &latticectl.SystemCommand{
		Name: "deploy",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "build",
				Required: true,
				Target:   &buildID,
			},
			&command.StringFlag{
				Name:     "version",
				Required: true,
				Target:   &version,
			},
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			systemID := ctx.SystemID()
			DeploySystem(ctx.Client().Systems().Rollouts(systemID), types.SystemBuildID(buildID), version)
		},
	}

	return cmd.Base()
}

func DeploySystem(
	client client.RolloutClient,
	buildID types.SystemBuildID,
	version string,
) {
	if buildID == "" && version == "" {
		log.Panic("must provide either build or version")
	}

	var deployID types.SystemRolloutID
	var err error
	if buildID != "" {
		if version != "" {
			log.Panic("can only provide either build or version")
			deployID, err = client.CreateFromBuild(buildID)
		}
	} else {
		deployID, err = client.CreateFromVersion(version)
	}

	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deployID)
}
