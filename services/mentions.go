package services

import (
	"regexp"
	"strings"
)

// StepPart is a piece of a recipe step: plain text, or a mention of one of the
// recipe's ingredients.
type StepPart struct {
	Text string
	// Ingredient is the index of the mentioned ingredient, or -1 for plain text.
	Ingredient int
}

// How surely a phrase names an ingredient. A step says "the ribs" for "bone-in
// short ribs" more often than "the chicken" for "chicken thighs", and a phrase
// several ingredients share only counts for the one that names it most surely.
const (
	byFirstWord = iota + 1
	byLastWords
	byWholeName
)

// vagueWords describe an ingredient without naming it, so they never name one alone.
var vagueWords = map[string]bool{
	"red": true, "green": true, "yellow": true, "white": true, "black": true, "brown": true,
	"orange": true, "purple": true, "dry": true, "ground": true, "whole": true, "fresh": true,
	"large": true, "small": true, "medium": true, "sweet": true, "hot": true, "cold": true,
	"extra": true, "virgin": true, "light": true, "dark": true, "boneless": true, "skinless": true,
	"bone-in": true, "skin-on": true, "unsalted": true, "salted": true, "kosher": true, "sea": true,
	"all-purpose": true, "plain": true, "low-sodium": true, "baby": true, "frozen": true,
	"canned": true, "raw": true, "cooked": true, "heavy": true, "lean": true, "warm": true,
	"and": true, "with": true, "for": true, "the": true, "of": true, "or": true,
}

var (
	wordRE         = regexp.MustCompile(`[\p{L}\p{N}]+(?:['’-][\p{L}\p{N}]+)*`)
	mentionFolding = strings.NewReplacer("ñ", "n", "é", "e", "è", "e", "’", "'")
)

// mentionWord normalizes a word so "Carrots" in a step matches "carrot" in a name.
func mentionWord(w string) string {
	return singularWord(mentionFolding.Replace(strings.ToLower(w)))
}

// StepParts splits a step into plain text and mentions of the given ingredients.
func StepParts(ingredients []string, step string) []StepPart {
	type claim struct{ strength, ingredient int }
	best := map[string]claim{}
	// Two ingredients claiming a phrase equally leaves it to neither.
	tied := map[string]bool{}
	longest := 0
	add := func(words []string, strength, ingredient int) {
		if len(words) == 1 && (len(words[0]) < 3 || vagueWords[words[0]]) {
			return
		}
		key := strings.Join(words, " ")
		c, ok := best[key]
		switch {
		case !ok || strength > c.strength:
			best[key] = claim{strength, ingredient}
			delete(tied, key)
		case strength == c.strength && ingredient != c.ingredient:
			tied[key] = true
		}
		longest = max(longest, len(words))
	}
	for i, line := range ingredients {
		for _, p := range parseIngredient(line) {
			var words []string
			for _, w := range wordRE.FindAllString(p.name, -1) {
				words = append(words, mentionWord(w))
			}
			if len(words) == 0 {
				continue
			}
			add(words, byWholeName, i)
			for k := 1; k < len(words); k++ {
				add(words[k:], byLastWords, i)
			}
			if len(words) > 1 && !strings.HasSuffix(words[0], "ed") {
				add(words[:1], byFirstWord, i)
			}
		}
	}

	spans := wordRE.FindAllStringIndex(step, -1)
	var parts []StepPart
	plainFrom := 0
	for i := 0; i < len(spans); {
		matched := 0
		ingredient := -1
		for n := min(longest, len(spans)-i); n > 0; n-- {
			if !adjacent(step, spans[i:i+n]) {
				continue
			}
			words := make([]string, n)
			for j, s := range spans[i : i+n] {
				words[j] = mentionWord(step[s[0]:s[1]])
			}
			key := strings.Join(words, " ")
			if c, ok := best[key]; ok && !tied[key] {
				matched, ingredient = n, c.ingredient
				break
			}
		}
		if matched == 0 {
			i++
			continue
		}
		start, end := spans[i][0], spans[i+matched-1][1]
		if start > plainFrom {
			parts = append(parts, StepPart{Text: step[plainFrom:start], Ingredient: -1})
		}
		parts = append(parts, StepPart{Text: step[start:end], Ingredient: ingredient})
		plainFrom = end
		i += matched
	}
	if plainFrom < len(step) {
		parts = append(parts, StepPart{Text: step[plainFrom:], Ingredient: -1})
	}
	return parts
}

// adjacent reports whether words are separated only by spaces, so a phrase
// doesn't run across punctuation as in "salt, olive oil".
func adjacent(text string, spans [][]int) bool {
	for j := 1; j < len(spans); j++ {
		if strings.TrimSpace(text[spans[j-1][1]:spans[j][0]]) != "" {
			return false
		}
	}
	return true
}
