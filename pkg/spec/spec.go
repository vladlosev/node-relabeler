package spec

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec contains a parsed relabel spec, ready to apply to node labels.
type Spec struct {
	OldKey   *regexp.Regexp
	OldValue *regexp.Regexp
	NewKey   string
	NewValue string
}

// ParseSpecs parses specs from the command line into format useful to apply
// them.
func ParseSpecs(specs []string) ([]Spec, error) {
	parsedSpecs := make([]Spec, 0, len(specs))
	for _, stringSpec := range specs {
		oldNew := strings.Split(stringSpec, ":")
		if len(oldNew) != 2 {
			return nil, newSpecParseError(stringSpec, "")
		}
		old := strings.Split(oldNew[0], "=")
		new := strings.Split(oldNew[1], "=")
		if len(old) > 2 || len(new) > 2 {
			return nil, newSpecParseError(stringSpec, "")
		}
		oldKey := old[0]
		oldValue := ""
		newKey := new[0]
		newValue := ""
		if len(old) == 2 {
			oldValue = old[1]
		}
		if len(new) == 2 {
			newValue = new[1]
		}
		if strings.Contains(oldKey, "*") && strings.Contains(oldValue, "*") {
			return nil, newSpecParseError(
				stringSpec,
				"oldkey=oldvalue pair should contain no more than a single *")
		}
		if strings.Contains(newKey, "*") && strings.Contains(newValue, "*") {
			return nil, newSpecParseError(
				stringSpec,
				"newkey=newvalue pair should contain no more than a single *")
		}
		if (strings.Contains(newKey, "*") || strings.Contains(newValue, "*")) &&
			!(strings.Contains(oldKey, "*") || strings.Contains(oldValue, "*")) {
			return nil, newSpecParseError(
				stringSpec,
				"Wildcard pattern can only appear in both old and new label")
		}

		parsedSpecs = append(parsedSpecs, Spec{
			OldKey: regexp.MustCompile(fmt.Sprintf(
				"^%s$",
				strings.Replace(oldKey, "*", "(.*)", 1),
			)),
			OldValue: regexp.MustCompile(fmt.Sprintf(
				"^%s$",
				strings.Replace(oldValue, "*", "(.*)", 1),
			)),
			NewKey:   strings.Replace(newKey, "*", "$1", 1),
			NewValue: strings.Replace(newValue, "*", "$1", 1),
		})
	}
	logrus.WithField("specs", parsedSpecs).Debug("Parsed specs from command line")
	return parsedSpecs, nil
}

func newSpecParseError(spec string, message string) error {
	if message == "" {
		message = "Specs must be in the form old/label=value:new/label=newvalue."
	}
	return fmt.Errorf("Invalid --relabel spec %s. %s", spec, message)
}
