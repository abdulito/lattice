package command

import (
	"fmt"
)

type Arg struct {
	Name     string
	Required bool
}

type Args []Arg

func (a Args) validate() error {
	var notRequired *string
	for _, arg := range a {
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
	for _, arg := range a {
		names = append(names, arg.Name)
	}

	return names
}

func (a Args) min() int {
	min := 0
	for _, arg := range a {
		if arg.Required {
			min++
		}
	}

	return min
}
