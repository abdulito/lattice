package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type GetCommand struct {
}

func (c *GetCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &latticectl.SystemCommand{
		Name: "get",
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			GetSystem(ctx.Client().Systems(), ctx.SystemID())
		},
	}

	return cmd.Base()
}

func GetSystem(client client.SystemClient, name types.SystemID) {
	system, err := client.Get(name)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}
