package mnemosyne

import "strings"

func Filter(skills []Skill, query string) []Skill {
	if query == "" {
		return skills
	}
	tokens := strings.Fields(strings.ToLower(query))
	var out []Skill
	for _, s := range skills {
		haystack := strings.ToLower(s.Name + " " + s.Description + " " + s.Category + " " + strings.Join(s.Tags, " "))
		match := true
		for _, t := range tokens {
			if !strings.Contains(haystack, t) {
				match = false
				break
			}
		}
		if match {
			out = append(out, s)
		}
	}
	return out
}
