package systems

import (
	"fmt"
	"log"

	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/latticectl/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

type DeployCommand struct {
}

func (c *DeployCommand) Base() (*latticectl.BaseCommand, error) {
	var buildID string
	var version string
	cmd := &command.SystemCommand{
		Name: "deploy",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "build",
				Required: false,
				Target:   &buildID,
			},
			&cli.StringFlag{
				Name:     "version",
				Required: false,
				Target:   &version,
			},
		},
		Run: func(ctx command.SystemCommandContext, args []string) {
			systemID := ctx.SystemID()
			DeploySystem(ctx.Client().Systems().Deploys(systemID), v1.BuildID(buildID), v1.SystemVersion(version))
		},
	}

	return cmd.Base()
}

func DeploySystem(
	client clientv1.DeployClient,
	buildID v1.BuildID,
	version v1.SystemVersion,
) {
	if buildID == "" && version == "" {
		log.Panic("must provide either build or version")
	}

	var deploy *v1.Deploy
	var err error
	if buildID != "" {
		if version != "" {
			log.Panic("can only provide either build or version")
			deploy, err = client.CreateFromBuild(buildID)
		}
	} else {
		deploy, err = client.CreateFromVersion(version)
	}

	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", deploy.ID)
}
