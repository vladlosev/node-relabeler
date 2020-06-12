package specs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec contains a parsed relabel spec, ready to apply to node labels.
type spec struct {
	oldKeyRegexp   *regexp.Regexp
	oldValueRegexp *regexp.Regexp
	oldKey         string
	oldValue       string
	newKey         string
	newValue       string
	stringSpec     string
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

		newSpec := spec{
			oldKeyRegexp: regexp.MustCompile(fmt.Sprintf(
				"^%s$",
				strings.Replace(oldKey, "*", "(.*)", 1),
			)),
			oldValueRegexp: regexp.MustCompile(fmt.Sprintf(
				"^%s$",
				strings.Replace(oldValue, "*", "(.*)", 1),
			)),
			oldKey:     oldKey,
			oldValue:   oldValue,
			newKey:     newKey,
			newValue:   newValue,
			stringSpec: stringSpec,
		}
		if strings.Contains(newSpec.oldKey, "*") &&
			strings.Contains(newSpec.newKey, "*") &&
			newSpec.oldKey != newSpec.newKey &&
			newSpec.oldKeyRegexp.MatchString(newSpec.newKey) {
			return nil, newSpecParseError(
				stringSpec,
				"newkey=newvalue pair must not match pattern in oldkey=oldvalue")
		}
		if strings.Contains(newSpec.oldValue, "*") &&
			strings.Contains(newSpec.newValue, "*") &&
			newSpec.oldValue != newSpec.newValue &&
			newSpec.oldKeyRegexp.MatchString(newSpec.newKey) &&
			newSpec.oldValueRegexp.MatchString(newSpec.newValue) {
			return nil, newSpecParseError(
				stringSpec,
				"newkey=newvalue pair must not match pattern in oldkey=oldvalue")
		}
		parsedSpecs = append(parsedSpecs, newSpec)
	}
	logrus.WithField("specs", parsedSpecs).Debug("Parsed specs from command line")
	return parsedSpecs, nil
}

// ApplyTo applies relabeling operations to a set of labels. Returns a map with
// changes to apply to the labels.
func (s Specs) ApplyTo(labels map[string]string) map[string]string {
	replacements := map[string]string{}

	for key, value := range labels {
		for _, spec := range s {
			keyMatch := spec.oldKeyRegexp.FindStringSubmatch(key)
			if keyMatch == nil {
				continue
			}
			valueMatch := spec.oldValueRegexp.FindStringSubmatch(value)
			if valueMatch == nil {
				continue
			}
			var newKey, newValue string
			if spec.oldKeyRegexp.NumSubexp() > 0 {
				newKey = strings.Replace(spec.newKey, "*", keyMatch[1], 1)
				newValue = strings.Replace(spec.newValue, "*", keyMatch[1], 1)
			} else if spec.oldValueRegexp.NumSubexp() > 0 {
				newKey = strings.Replace(spec.newKey, "*", valueMatch[1], 1)
				newValue = strings.Replace(spec.newValue, "*", valueMatch[1], 1)

			} else {
				newKey = spec.newKey
				newValue = spec.newValue
			}
			replacements[newKey] = newValue
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
