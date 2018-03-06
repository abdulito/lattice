package latticectl

import (
	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeployCommand struct {
	Name        string
	Short       string
	Args        command.Args
	Flags       command.Flags
	Run         func(ctx DeployCommandContext, args []string)
	Subcommands []Command
}

type DeployCommandContext interface {
	SystemCommandContext
	DeployID() types.SystemRolloutID
}

type deployCommandContext struct {
	SystemCommandContext
	deployID types.SystemRolloutID
}

func (c *deployCommandContext) DeployID() types.SystemRolloutID {
	return c.deployID
}

func (c *DeployCommand) Base() (*BaseCommand, error) {
	var deployID string
	deployIDFlag := &command.StringFlag{
		Name:     "deploy",
		Required: true,
		Target:   &deployID,
	}
	flags := append(c.Flags, deployIDFlag)

	cmd := &SystemCommand{
		Name:  c.Name,
		Short: c.Short,
		Args:  c.Args,
		Flags: flags,
		Run: func(sctx SystemCommandContext, args []string) {
			ctx := &deployCommandContext{
				SystemCommandContext: sctx,
				deployID:             types.SystemRolloutID(deployID),
			}
			c.Run(ctx, args)
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}
