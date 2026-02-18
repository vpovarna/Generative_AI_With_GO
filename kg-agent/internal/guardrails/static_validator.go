package guardrails

import (
	"fmt"
	"regexp"
	"strings"
)

type StaticValidator struct {
	banWords []string
}

func NewStaticValidator(banWords []string) *StaticValidator {
	return &StaticValidator{
		banWords: banWords,
	}
}

func (v *StaticValidator) Validate(input string) ValidationResult {
	// Check for ban words
	lowerInput := strings.ToLower(input)
	for _, word := range v.banWords {
		if v.containsBanWord(lowerInput, word) {
			return ValidationResult{
				IsValid:  false,
				Reason:   fmt.Sprintf("Contains banned word: %s", word),
				Category: "banned_word",
				Method:   "static",
			}
		}
	}

	return ValidationResult{IsValid: true, Method: "static"}
}

func (v *StaticValidator) containsBanWord(input string, banWord string) bool {
	lowerBanWord := strings.ToLower(banWord)

	// Pattern: \bword\b means "word" as a complete word (avoids false positives)
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(lowerBanWord))
	matched, _ := regexp.MatchString(pattern, input)

	return matched
}
