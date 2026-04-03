package prtype

import (
	"strings"
)

// TypeConfig holds mapping from type prefix to emoji
type TypeConfig struct {
	Emoji string
}

var (
	// TypeMapping is the source of truth for all supported prefixes and their emojis.
	TypeMapping = map[string]TypeConfig{
		"migrate":   {Emoji: "🧹"},
		"migration": {Emoji: "🧹"},
		"chore":     {Emoji: "🧹"},
		"feat":      {Emoji: "✨"},
		"feature":   {Emoji: "✨"},
		"terraform": {Emoji: "🚜"},
		"tf":        {Emoji: "🚜"},
		"infra":     {Emoji: "🚜"},
		"bug":       {Emoji: "🐛"},
		"fix":       {Emoji: "🐛"},
		"hotfix":    {Emoji: "🐛"},
		"bugfix":    {Emoji: "🐛"},
		"polish":    {Emoji: "💄"},
		"docs":      {Emoji: "💄"},
		"doc":       {Emoji: "💄"},
	}

	// SortedTypes defines the order and characters to check for branch prefixes.
	// We want to match longer prefixes first to avoid partial matches (e.g., "bugfix" before "bug").
	SortedTypes []string
)

func init() {
	// Simple sorting by length descending to ensure longer matches win
	SortedTypes = make([]string, 0, len(TypeMapping))
	for t := range TypeMapping {
		SortedTypes = append(SortedTypes, t)
	}

	for i := 0; i < len(SortedTypes); i++ {
		for j := i + 1; j < len(SortedTypes); j++ {
			if len(SortedTypes[i]) < len(SortedTypes[j]) {
				SortedTypes[i], SortedTypes[j] = SortedTypes[j], SortedTypes[i]
			}
		}
	}
}

// DetectTypeFromBranch returns the normalized type from a branch name.
func DetectTypeFromBranch(branchName string) string {
	lowerBranch := strings.ToLower(branchName)
	for _, p := range SortedTypes {
		if strings.HasPrefix(lowerBranch, p+"/") || strings.HasPrefix(lowerBranch, p+"-") {
			return p
		}
	}
	return ""
}

// StripPrefix removes recognized prefixes from a branch name.
func StripPrefix(branchName string) string {
	lowerBranch := strings.ToLower(branchName)
	for _, p := range SortedTypes {
		if strings.HasPrefix(lowerBranch, p+"/") || strings.HasPrefix(lowerBranch, p+"-") {
			prefixSlash := p + "/"
			if strings.HasPrefix(lowerBranch, prefixSlash) {
				return branchName[len(prefixSlash):]
			}
			prefixDash := p + "-"
			if strings.HasPrefix(lowerBranch, prefixDash) {
				return branchName[len(prefixDash):]
			}
		}
	}
	return branchName
}

// InferTypeFromTitle searches for keywords in the title and returns the matched type.
func InferTypeFromTitle(title string) string {
	if title == "" {
		return ""
	}
	lowerTitle := strings.ToLower(title)
	for _, p := range SortedTypes {
		if strings.Contains(lowerTitle, p) {
			return p
		}
	}
	return ""
}

// GetEmoji returns the emoji for a given type, defaulting to "✨" for "feat" if unknown.
func GetEmoji(typeName string) string {
	if cfg, ok := TypeMapping[strings.ToLower(typeName)]; ok {
		return cfg.Emoji
	}
	return "✨"
}
