package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var OnInitialize = cobra.OnInitialize

type Command interface {
	Init() error
	name() string
	run() func(args []string)
	subcommands() []Command
	cobra() *cobra.Command
}

type BaseCommand struct {
	Name        string
	Short       string
	Args        Args
	Flags       Flags
	PreRun      func()
	Run         func(args []string)
	Subcommands []Command
	cobraCmd    *cobra.Command
}

func (c *BaseCommand) Execute() {
	if err := c.validate(); err != nil {
		c.exit(err)
	}

	if err := c.Init(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *BaseCommand) Init() error {
	c.cobraCmd = &cobra.Command{
		Use:   c.Name,
		Short: c.Short,
		Run: func(cmd *cobra.Command, args []string) {
			if c.Run == nil {
				cmd.Help()
				os.Exit(1)
			}
			c.Run(args)
		},
	}

	if err := c.addArgs(); err != nil {
		return fmt.Errorf("error adding args: %v", err)
	}

	if err := c.addFlags(); err != nil {
		return fmt.Errorf("error adding flags: %v", err)
	}

	if err := c.addSubcommands(); err != nil {
		return fmt.Errorf("error initializing subcommands: %v", err)
	}

	return nil
}

func (c *BaseCommand) name() string {
	return c.Name
}

func (c *BaseCommand) run() func([]string) {
	return c.Run
}

func (c *BaseCommand) subcommands() []Command {
	return c.Subcommands
}

func (c *BaseCommand) cobra() *cobra.Command {
	return c.cobraCmd
}

func (c *BaseCommand) validate() error {
	if c.Name == "" {
		return fmt.Errorf("name must be set")
	}

	return nil
}

func (c *BaseCommand) addArgs() error {
	if err := c.Args.validate(); err != nil {
		return err
	}

	for _, arg := range c.Args {
		c.cobraCmd.Use += fmt.Sprintf(" [%v]", arg.Name)
	}

	c.cobraCmd.Args = cobra.MinimumNArgs(c.Args.min())

	return nil
}

func (c *BaseCommand) addFlags() error {
	names := make(map[string]struct{})
	for _, flag := range c.Flags {
		if err := flag.validate(); err != nil {
			return fmt.Errorf("error validating flag %v: %v", flag.name(), err)
		}

		if _, ok := names[flag.name()]; ok {
			return fmt.Errorf("multiple flags with the name %v", flag.name())
		}

		flag.addToCmd(c.cobraCmd)
		names[flag.name()] = struct{}{}
	}

	return nil
}

func (c *BaseCommand) addSubcommands() error {
	names := make(map[string]struct{})
	for _, subcommand := range c.Subcommands {
		if err := subcommand.Init(); err != nil {
			return err
		}

		if _, ok := names[subcommand.name()]; ok {
			return fmt.Errorf("multiple subcommands with the name %v", c.Name)
		}

		if err := subcommand.Init(); err != nil {
			return fmt.Errorf("error initializing subcommand %v: %v", c.Name, err)
		}

		c.cobraCmd.AddCommand(subcommand.cobra())
		names[subcommand.name()] = struct{}{}
	}

	return nil
}

func (c *BaseCommand) ExecuteColon() {
	if err := c.validate(); err != nil {
		c.exit(err)
	}

	if err := c.initColon(); err != nil {
		c.exit(err)
	}

	c.exit(c.cobraCmd.Execute())
}

func (c *BaseCommand) initColon() error {
	c.cobraCmd = &cobra.Command{
		Use:   c.Name,
		Short: c.Short,
		Run: func(cmd *cobra.Command, args []string) {
			if c.Run == nil {
				cmd.Help()
				os.Exit(1)
			}

			c.Run(args)
		},
	}

	for _, subcommand := range c.Subcommands {
		if err := subcommand.Init(); err != nil {
			return err
		}
	}

	for _, subcommand := range getSubcommands("", c.Subcommands) {
		// why does this need to be an immediately invoked function?
		// answer here: https://www.ardanlabs.com/blog/2014/06/pitfalls-with-closures-in-go.html
		// (n.b. subcommand.Name will be copied here since it's a string, but since
		//  subcommand.Run is a pointer, we need to do this trickery)
		subcommand.cobra().Run = func(run func([]string)) func(*cobra.Command, []string) {
			return func(cmd *cobra.Command, args []string) {
				run(args)
			}
		}(subcommand.run())

		c.cobraCmd.AddCommand(subcommand.cobra())
	}

	return nil
}

func getSubcommands(path string, subcommands []Command) []Command {
	var ret []Command
	for _, subcommand := range subcommands {
		name := fmt.Sprintf("%v%v", path, subcommand.name())
		subcommand.cobra().Use = name
		ret = append(ret, subcommand)
		for _, subsubcommand := range getSubcommands(fmt.Sprintf("%v:", name), subcommand.subcommands()) {
			ret = append(ret, subsubcommand)
		}

	}

	return ret
}

func (c *BaseCommand) exit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
