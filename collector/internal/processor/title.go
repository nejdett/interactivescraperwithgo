package processor

import (
	"regexp"
	"strings"
	"unicode"
)

type TitleGenerator interface {
	Generate(content string) string
}

type RuleBasedGenerator struct {
	maxLength int
}

func NewRuleBasedGenerator() *RuleBasedGenerator {
	return &RuleBasedGenerator{
		maxLength: 100,
	}
}

func (rbg *RuleBasedGenerator) Generate(content string) string {
	if content == "" {
		return "Untitled Content"
	}

	cleaned := rbg.cleanContent(content)
	firstSentence := rbg.extractFirstSentence(cleaned)
	title := rbg.applyLengthLimit(firstSentence)
	title = strings.TrimSpace(title)
	
	if len(title) < 10 {
		return rbg.fallbackTitle(content)
	}
	
	return title
}

func (rbg *RuleBasedGenerator) cleanContent(content string) string {
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = regexp.MustCompile(`[^\w\s\.\,\-\:\;\!\?]`).ReplaceAllString(content, "")
	return strings.TrimSpace(content)
}

func (rbg *RuleBasedGenerator) extractFirstSentence(content string) string {
	enders := []string{". ", "! ", "? "}
	
	minIdx := len(content)
	for _, e := range enders {
		if idx := strings.Index(content, e); idx != -1 && idx < minIdx {
			minIdx = idx
		}
	}
	
	if minIdx < len(content) {
		return content[:minIdx]
	}
	return content
}

func (rbg *RuleBasedGenerator) applyLengthLimit(text string) string {
	if len(text) <= rbg.maxLength {
		return text
	}
	
	truncated := text[:rbg.maxLength]
	lastSpace := strings.LastIndexFunc(truncated, unicode.IsSpace)
	
	if lastSpace > 0 {
		return truncated[:lastSpace]
	}
	return truncated
}

func (rbg *RuleBasedGenerator) fallbackTitle(content string) string {
	if len(content) <= rbg.maxLength {
		return strings.TrimSpace(content)
	}
	
	truncated := content[:rbg.maxLength]
	lastSpace := strings.LastIndexFunc(truncated, unicode.IsSpace)
	
	if lastSpace > 0 {
		return strings.TrimSpace(truncated[:lastSpace]) + "..."
	}
	return strings.TrimSpace(truncated) + "..."
}
