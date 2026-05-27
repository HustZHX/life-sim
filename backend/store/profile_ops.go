package store

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"life-sim/backend/model"
)

var profileRandomizableFields = map[string]bool{
	"display_name": true, "birth_year": true, "death_year": true,
	"era": true, "era_background": true, "personality_initial": true,
	"beliefs_motto": true, "experiences": true, "literature": true,
	"birth_place": true, "nationality": true, "social_class": true,
	"occupation": true, "education": true, "family_relations": true,
	"beliefs_politics": true, "appearance": true, "major_works": true,
	"controversies": true, "death_cause": true, "sources_note": true,
}

func ValidProfileField(field string) bool {
	return profileRandomizableFields[field]
}

func MergeProfilePatch(dst, src *model.Profile) {
	if src.DisplayName != "" {
		dst.DisplayName = src.DisplayName
	}
	if src.BirthYear != 0 {
		dst.BirthYear = src.BirthYear
	}
	if src.DeathYear != 0 {
		dst.DeathYear = src.DeathYear
	}
	if src.Era != "" {
		dst.Era = src.Era
	}
	if src.EraBackground != "" {
		dst.EraBackground = src.EraBackground
	}
	if src.PersonalityInitial != "" {
		dst.PersonalityInitial = src.PersonalityInitial
	}
	if src.BeliefsMotto != "" {
		dst.BeliefsMotto = src.BeliefsMotto
	}
	if src.Experiences != "" {
		dst.Experiences = src.Experiences
	}
	if src.Literature != "" {
		dst.Literature = src.Literature
	}
	if src.BirthPlace != "" {
		dst.BirthPlace = src.BirthPlace
	}
	if src.Nationality != "" {
		dst.Nationality = src.Nationality
	}
	if src.SocialClass != "" {
		dst.SocialClass = src.SocialClass
	}
	if src.Occupation != "" {
		dst.Occupation = src.Occupation
	}
	if src.Education != "" {
		dst.Education = src.Education
	}
	if src.FamilyRelations != "" {
		dst.FamilyRelations = src.FamilyRelations
	}
	if src.BeliefsPolitics != "" {
		dst.BeliefsPolitics = src.BeliefsPolitics
	}
	if src.Appearance != "" {
		dst.Appearance = src.Appearance
	}
	if src.MajorWorks != "" {
		dst.MajorWorks = src.MajorWorks
	}
	if src.Controversies != "" {
		dst.Controversies = src.Controversies
	}
	if src.DeathCause != "" {
		dst.DeathCause = src.DeathCause
	}
	if src.SourcesNote != "" {
		dst.SourcesNote = src.SourcesNote
	}
}

func ApplyProfileFieldValue(p *model.Profile, field string, raw json.RawMessage) error {
	switch field {
	case "birth_year", "death_year":
		var n int
		if err := json.Unmarshal(raw, &n); err != nil {
			var s string
			if err2 := json.Unmarshal(raw, &s); err2 != nil {
				return err
			}
			v, err3 := strconv.Atoi(s)
			if err3 != nil {
				return err3
			}
			n = v
		}
		if field == "birth_year" {
			p.BirthYear = n
		} else {
			p.DeathYear = n
		}
		return nil
	default:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return err
		}
		switch field {
		case "display_name":
			p.DisplayName = s
		case "era":
			p.Era = s
		case "era_background":
			p.EraBackground = s
		case "personality_initial":
			p.PersonalityInitial = s
		case "beliefs_motto":
			p.BeliefsMotto = s
		case "experiences":
			p.Experiences = s
		case "literature":
			p.Literature = s
		case "birth_place":
			p.BirthPlace = s
		case "nationality":
			p.Nationality = s
		case "social_class":
			p.SocialClass = s
		case "occupation":
			p.Occupation = s
		case "education":
			p.Education = s
		case "family_relations":
			p.FamilyRelations = s
		case "beliefs_politics":
			p.BeliefsPolitics = s
		case "appearance":
			p.Appearance = s
		case "major_works":
			p.MajorWorks = s
		case "controversies":
			p.Controversies = s
		case "death_cause":
			p.DeathCause = s
		case "sources_note":
			p.SourcesNote = s
		default:
			return fmt.Errorf("未知字段: %s", field)
		}
		return nil
	}
}

func ParseSuggestNamesResult(raw string) ([]string, error) {
	var resp struct {
		Names []string `json:"names"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	if len(resp.Names) == 0 {
		return nil, fmt.Errorf("AI 未返回姓名列表")
	}
	seen := map[string]bool{}
	out := make([]string, 0, 5)
	for _, n := range resp.Names {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
		if len(out) >= 5 {
			break
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("无有效姓名")
	}
	return out, nil
}

func ParseRandomizeFieldResult(raw, field string) (json.RawMessage, error) {
	var resp struct {
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	if len(resp.Value) == 0 {
		return nil, fmt.Errorf("AI 未返回 value")
	}
	return resp.Value, nil
}
