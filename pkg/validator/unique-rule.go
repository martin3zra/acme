package validator

import "strings"

func (r Rule) Unique() *Unique {
	return new(Unique)
}

type Unique struct {
	constraints []string
}

func (u Unique) Constraints() string {
	var rules = make([]string, 0)

	rules = append(rules, u.constraints...)

	return strings.Join(rules, "|")
}
