package context

import (
	"github.com/mlab-lattice/lattice/pkg/latticectl2/command"
	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/printer"
	"io"
	"os"
)

func Command() *cli.Command {
	var (
		configPath string
		output     string
	)

	return &cli.Command{
		Flags: cli.Flags{
			command.ConfigFlagName: command.ConfigFlag(&configPath),
			command.OutputFlagName: command.OutputFlag(&output, []printer.Format{printer.FormatJSON}, printer.FormatJSON),
		},
		Run: func(args []string, flags cli.Flags) error {
			// if ConfigFile.Path is empty, it will look in $XDG_CONFIG_HOME/.latticectl/config.json
			configFile := command.ConfigFile{Path: configPath}

			contextName, err := configFile.CurrentContext()
			if err != nil {
				return err
			}

			context, err := configFile.Context(contextName)
			if err != nil {
				return err
			}

			format := printer.Format(output)
			return PrintContext(context, os.Stdout, format)
		},
		Subcommands: map[string]*cli.Command{
			"create": Create(),
			"delete": Delete(),
			"list":   List(),
			"switch": Switch(),
			"update": Update(),
		},
	}
}

func PrintContext(ctx *command.Context, w io.Writer, f printer.Format) error {
	// FIXME: probably want to support a more natural table-like format
	switch f {
	case printer.FormatJSON:
		j := printer.NewJSONIndented(os.Stdout, 4)
		j.Print(ctx)
	}

	return nil
}
