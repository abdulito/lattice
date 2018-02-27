package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/types"
)

type DeleteCommand struct {
	PreRun         func()
	ContextCreator func(ctx latticectl.LatticeCommandContext, systemID types.SystemID) latticectl.SystemCommandContext
	*latticectl.SystemCommand
}

func (c *DeleteCommand) Init() error {
	if c.ContextCreator == nil {
		c.ContextCreator = latticectl.DefaultSystemContextCreator
	}

	var definitionURL string
	var systemName string
	c.SystemCommand = &latticectl.SystemCommand{
		Name: "delete",
		Run: func(args []string, ctx latticectl.SystemCommandContext) {
			c.run(ctx, types.SystemID(systemName), definitionURL)
		},
		ContextCreator: c.ContextCreator,
	}

	return c.SystemCommand.Init()
}

func (c *DeleteCommand) run(ctx latticectl.SystemCommandContext, name types.SystemID, definitionURL string) {
	system, err := ctx.Client().Systems().Get(ctx.SystemID())
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}
