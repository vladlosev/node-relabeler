package specs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec contains a parsed relabel spec, ready to apply to node labels.
type spec struct {
	OldKey   *regexp.Regexp
	OldValue *regexp.Regexp
	NewKey   string
	NewValue string
}

// Specs keeps compiled relabeling specs and applies them.
type Specs []spec

// Parse parses specs from the command line into format useful to apply
// them.
func Parse(specs []string) (Specs, error) {
	parsedSpecs := make([]spec, 0, len(specs))
	if len(specs) == 0 {
		return nil, fmt.Errorf("At least one --relabel spec must be specified")
	}
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
		if (strings.Contains(newKey, "*") || strings.Contains(newValue, "*")) &&
			!(strings.Contains(oldKey, "*") || strings.Contains(oldValue, "*")) {
			return nil, newSpecParseError(
				stringSpec,
				"Wildcard pattern cannot appear in new label without appearing in the old one")
		}

		parsedSpecs = append(parsedSpecs, spec{
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

// ApplyTo applies relabeling operations to a set of labels. Returns a map with
// changes to apply to the labels.
func (s Specs) ApplyTo(labels map[string]string) map[string]string {
	var replacements map[string]string

	for key, value := range labels {
		for _, spec := range s {
			if spec.OldKey.MatchString(key) && spec.OldValue.MatchString(value) {
				var newKey, newValue string
				if spec.OldKey.NumSubexp() > 0 {
					newKey = spec.OldKey.ReplaceAllString(key, spec.NewKey)
					newValue = spec.OldKey.ReplaceAllString(value, spec.NewValue)
				} else if spec.OldValue.NumSubexp() > 0 {
					newKey = spec.OldValue.ReplaceAllString(key, spec.NewKey)
					newValue = spec.OldValue.ReplaceAllString(value, spec.NewValue)
				} else {
					newKey = spec.NewKey
					newValue = spec.NewValue
				}
				replacements[newKey] = newValue
			}
		}
	}
	return replacements
}

func newSpecParseError(spec string, message string) error {
	if message == "" {
		message = "Specs must be in the form old/label=value:new/label=newvalue."
	}
	return fmt.Errorf("Invalid --relabel spec %s. %s", spec, message)
}
