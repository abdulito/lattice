package systems

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type CreateCommand struct {
}

func (c *CreateCommand) Base() (*latticectl.BaseCommand, error) {
	var definitionURL string
	var systemName string
	cmd := &latticectl.LatticeCommand{
		Name: "create",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:     "definition",
				Required: true,
				Target:   &definitionURL,
			},
			&command.StringFlag{
				Name:     "name",
				Required: true,
				Target:   &systemName,
			},
		},
		Run: func(ctx latticectl.LatticeCommandContext, args []string) {
			CreateSystem(ctx.Client().Systems(), types.SystemID(systemName), definitionURL)
		},
	}

	return cmd.Base()
}

func CreateSystem(client client.SystemClient, name types.SystemID, definitionURL string) {
	system, err := client.Create(name, definitionURL)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%v\n", system)
}
