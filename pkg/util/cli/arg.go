package cli

import (
	"fmt"
)

type Arg struct {
	Name        string
	Required    bool
	Description string
}

type Args struct {
	Args            []Arg
	AllowAdditional bool
}

func (a Args) validate() error {
	var notRequired *string
	for _, arg := range a.Args {
		if !arg.Required {
			notRequired = &arg.Name
			continue
		}

		if arg.Required && notRequired != nil {
			return fmt.Errorf(
				"error parsing arguments: argument %v required after argument %v was not",
				arg.Name,
				notRequired,
			)
		}
	}

	return nil
}

func (a Args) names() []string {
	var names []string
	for _, arg := range a.Args {
		names = append(names, arg.Name)
	}

	return names
}

func (a Args) num() (int, int) {
	min := 0
	max := 0
	for _, arg := range a.Args {
		if arg.Required {
			min++
		}
		max++
	}

	return min, max
}
