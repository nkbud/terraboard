package compare

import (
	"fmt"
	"sort"
	"strings"

	"github.com/camptocamp/terraboard/types"
	"github.com/pmezard/go-difflib/difflib"
	log "github.com/sirupsen/logrus"
)

// Return all resources of a state
func stateResources(state types.State) (res []string) {
	for _, m := range state.Modules {
		for _, r := range m.Resources {
			res = append(res, fmt.Sprintf("%s.%s.%s", m.Path, r.Type, r.Name))
		}
	}
	return
}

// Returns elements only in s1
func sliceDiff(s1, s2 []string) (diff []string) {
	for _, e1 := range s1 {
		found := false
		for _, e2 := range s2 {
			if e1 == e2 {
				found = true
				break
			}
		}

		if !found {
			diff = append(diff, e1)
		}
	}
	return
}

// Returns elements in both s1 and s2
func sliceInter(s1, s2 []string) (inter []string) {
	for _, e1 := range s1 {
		for _, e2 := range s2 {
			if e1 == e2 {
				inter = append(inter, e1)
				break
			}
		}
	}
	return
}

func getResource(state types.State, key string) (res types.Resource, err error) {
	for _, m := range state.Modules {
		if strings.HasPrefix(key, m.Path) {
			for _, r := range m.Resources {
				if key == fmt.Sprintf("%s.%s.%s", m.Path, r.Type, r.Name) {
					return r, nil
				}
			}
		} else {
			continue
		}
	}
	return res, fmt.Errorf("Could not find resource with key %s in state %s", key, state.Path)
}

// Return all attributes of a resource
func resourceAttributes(res types.Resource) (attrs []string) {
	for _, a := range res.Attributes {
		attrs = append(attrs, a.Key)
	}
	sort.Strings(attrs)
	return
}

func getResourceAttribute(res types.Resource, key string) (attr types.Attribute, err error) {
	for _, a := range res.Attributes {
		if a.Key == key {
			return a, nil
		}
	}
	return attr, fmt.Errorf("could not find attribute %s for resource %s.%s", key, res.Type, res.Name)
}

// TODO: use terraform/command/format.State()
func formatResource(res types.Resource) (out string) {
	out = fmt.Sprintf("resource \"%s\" \"%s\" {\n", res.Type, res.Name)
	for _, attrKey := range resourceAttributes(res) {
		attr, _ := getResourceAttribute(res, attrKey) // TODO: err
		if attr.Sensitive {
			if attr.Value == "null" {
				out += fmt.Sprintf("  %s = (null)\n", attr.Key)
			} else {
				out += fmt.Sprintf("  %s = (%d)\n", attr.Key, len(attr.Value))
			}
		} else {
			out += fmt.Sprintf("  %s = %s\n", attr.Key, attr.Value)
		}
	}
	out += "}\n"

	return
}

func stateInfo(state types.State) (info string) {
	return fmt.Sprintf("%s (%s)", state.Path, state.Version.LastModified)
}

// Compare a resource in two states
func compareResource(st1, st2 types.State, key string) (comp types.ResourceDiff) {
	res1, _ := getResource(st1, key) // TODO: err
	attrs1 := resourceAttributes(res1)
	res2, _ := getResource(st2, key) // TODO: err
	attrs2 := resourceAttributes(res2)

	// Only in old
	comp.OnlyInOld = make(map[string]string)
	for _, attrKey := range sliceDiff(attrs1, attrs2) {
		attr, _ := getResourceAttribute(res1, attrKey) // TODO: err
		if attr.Sensitive {
			if attr.Value == "null" {
				comp.OnlyInOld[attr.Key] = "(null)"
			} else {
				comp.OnlyInOld[attr.Key] = fmt.Sprintf("(%d)", len(attr.Value))
			}
		} else {
			comp.OnlyInOld[attr.Key] = attr.Value
		}
	}

	// Only in new
	comp.OnlyInNew = make(map[string]string)
	for _, attrKey := range sliceDiff(attrs2, attrs1) {
		attr, _ := getResourceAttribute(res2, attrKey) // TODO: err
		if attr.Sensitive {
			if attr.Value == "null" {
				comp.OnlyInNew[attr.Key] = "(null)"
			} else {
				comp.OnlyInNew[attr.Key] = fmt.Sprintf("(%d)", len(attr.Value))
			}
		} else {
			comp.OnlyInNew[attr.Key] = attr.Value
		}
	}

	// Compute unified diff
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(formatResource(res1)),
		B:        difflib.SplitLines(formatResource(res2)),
		FromFile: stateInfo(st1),
		ToFile:   stateInfo(st2),
		Context:  3,
		Eol:      "\n",
	}
	result, _ := difflib.GetUnifiedDiffString(diff)
	comp.UnifiedDiff = result

	return
}

// Compare returns the difference between two versions of a State
// as a StateCompare structure
func Compare(from, to types.State) (comp types.StateCompare, err error) {
	if from.Path == "" {
		err = fmt.Errorf("from version is unknown")
		return
	}
	fromResources := stateResources(from)
	comp.Stats.From = types.StateInfo{
		Path:          from.Path,
		VersionID:     from.Version.VersionID,
		ResourceCount: len(fromResources),
		TFVersion:     from.TFVersion,
		Serial:        from.Serial,
	}

	if to.Path == "" {
		err = fmt.Errorf("to version is unknown")
		return
	}
	toResources := stateResources(to)
	comp.Stats.To = types.StateInfo{
		Path:          to.Path,
		VersionID:     to.Version.VersionID,
		ResourceCount: len(toResources),
		TFVersion:     to.TFVersion,
		Serial:        to.Serial,
	}

	// OnlyInOld
	onlyInOld := sliceDiff(fromResources, toResources)
	comp.Differences.OnlyInOld = make(map[string]string)
	for _, r := range onlyInOld {
		res, _ := getResource(from, r) // TODO: err
		comp.Differences.OnlyInOld[r] = formatResource(res)
	}

	// OnlyInNew
	onlyInNew := sliceDiff(toResources, fromResources)
	comp.Differences.OnlyInNew = make(map[string]string)
	for _, r := range onlyInNew {
		res, _ := getResource(to, r) // TODO: err
		comp.Differences.OnlyInNew[r] = formatResource(res)
	}
	comp.Differences.InBoth = sliceInter(toResources, fromResources)
	comp.Differences.ResourceDiff = make(map[string]types.ResourceDiff)

	for _, r := range comp.Differences.InBoth {
		if c := compareResource(to, from, r); c.UnifiedDiff != "" {
			comp.Differences.ResourceDiff[r] = c
		}
	}

	log.WithFields(log.Fields{
		"path": from.Path,
		"from": from.Version.VersionID,
		"to":   to.Version.VersionID,
	}).Info("Comparing state versions")

	return
}
