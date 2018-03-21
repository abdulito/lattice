package command

import (
	"fmt"

	"github.com/spf13/pflag"
)

type StringFlag struct {
	Name     string
	Required bool
	Default  string
	Short    string
	Usage    string
	Target   *string
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) IsRequired() bool {
	return f.Required
}

func (f *StringFlag) GetShort() string {
	return f.Short
}

func (f *StringFlag) GetUsage() string {
	return f.Usage
}

func (f *StringFlag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be nil")
	}

	if f.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	return nil
}

func (f *StringFlag) GetTarget() interface{} {
	return f.Target
}

func (f *StringFlag) Parse() func() error {
	return nil
}

func (f *StringFlag) AddToFlagSet(flags *pflag.FlagSet) {
	flags.StringVarP(f.Target, f.Name, f.Short, f.Default, f.Usage)

	if f.Required {
		markFlagRequired(f.Name, flags)
	}
}
