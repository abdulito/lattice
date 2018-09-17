package command

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/client"
	"github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
)

type LatticeCommandContext struct {
	Context *Context
	Lattice string
	Client  client.Interface
}

type LatticeCommand struct {
	Short       string
	Args        cli.Args
	Flags       cli.Flags
	Run         func(ctx *LatticeCommandContext, args []string, flags cli.Flags) error
	Subcommands map[string]*cli.Command
}

func (c *LatticeCommand) Command() *cli.Command {
	c.Flags[ConfigFlagName] = ConfigFlag()
	c.Flags[ContextFlagName] = ContextFlag()

	cmd := &cli.Command{
		Short: c.Short,
		Args:  c.Args,
		Flags: c.Flags,
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configPath := c.Flags[ConfigFlagName].Value().(string)
			configFile := ConfigFile{Path: configPath}

			contextName := c.Flags[ContextFlagName].Value().(string)
			if contextName == "" {
				var err error
				contextName, err = configFile.CurrentContext()
				if err != nil {
					return err
				}
			}

			context, err := configFile.Context(contextName)
			if err != nil {
				return err
			}

			var client client.Interface
			switch {
			case context.Auth == nil:
				client = rest.NewUnauthenticatedClient(context.Lattice)

			case context.Auth.BearerToken != nil:
				client = rest.NewBearerTokenClient(context.Lattice, *context.Auth.BearerToken)

			default:
				return fmt.Errorf("invalid auth options for context %v", contextName)
			}

			ctx := &LatticeCommandContext{
				Context: context,
				Client:  client,
			}
			return c.Run(ctx, args, flags)
		},
		Subcommands: c.Subcommands,
	}

	return cmd
}
