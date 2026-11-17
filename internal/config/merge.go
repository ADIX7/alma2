package config

import (
	"sort"
	"strings"
)

// Settings are the effective values for one module and platform identifier.
// A nil module is unconfigured and remains linkable without README exclusion;
// a configured module with no matching sections is unavailable.
type Settings struct {
	Available     bool
	Target        *string
	Links         []Link
	Exclude       []string
	ExcludeReadme bool
	Install       *string
	Configure     *string
}

// Select merges matching sections from least to most specific. Equal-length
// selectors retain YAML declaration order. Duplicate sources retain their
// first position while their later mapping value replaces the earlier one.
func Select(module *Module, identifier string) Settings {
	if module == nil {
		return Settings{Available: true}
	}
	settings := Settings{ExcludeReadme: true}
	matching := make([]Platform, 0, len(module.Platforms))
	for _, section := range module.Platforms {
		if section.Selector == "default" || strings.HasPrefix(strings.ToLower(identifier), strings.ToLower(section.Selector)) {
			matching = append(matching, section)
		}
	}
	if len(matching) == 0 {
		return settings
	}
	settings.Available = true
	sort.SliceStable(matching, func(i, j int) bool {
		if matching[i].Selector == "default" {
			return matching[j].Selector != "default"
		}
		if matching[j].Selector == "default" {
			return false
		}
		return len(matching[i].Selector) < len(matching[j].Selector)
	})
	linkPositions := make(map[string]int)
	for _, section := range matching {
		if section.Target != nil {
			settings.Target = section.Target
		}
		if section.Install != nil {
			settings.Install = section.Install
		}
		if section.Configure != nil {
			settings.Configure = section.Configure
		}
		if section.ExcludeReadme != nil {
			settings.ExcludeReadme = *section.ExcludeReadme
		}
		settings.Exclude = append(settings.Exclude, section.Exclude...)
		for _, link := range section.Links {
			if pos, exists := linkPositions[link.Source]; exists {
				settings.Links[pos] = link
			} else {
				linkPositions[link.Source] = len(settings.Links)
				settings.Links = append(settings.Links, link)
			}
		}
	}
	return settings
}
