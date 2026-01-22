package processor

import (
	"strings"
)

type CriticalityScorer struct {
	keywords         map[string]int
	categoryModifiers map[string]int
}

func NewCriticalityScorer() *CriticalityScorer {
	return &CriticalityScorer{
		keywords: map[string]int{
			"ransomware": 10, "zero-day": 10, "zero day": 10,
			"data breach": 9, "apt": 9, "advanced persistent threat": 9,
			"exploit": 8, "remote code execution": 8, "rce": 8, "backdoor": 8,
			"credential": 7, "stolen data": 7,
			"vulnerability": 6, "malware": 6, "trojan": 6,
			"phishing": 5, "botnet": 5,
			"suspicious": 4, "attack": 4, "compromise": 4,
			"threat": 3, "risk": 3,
		},
		categoryModifiers: map[string]int{
			"ransomware": 2, "data-leak": 2, "exploit": 1,
			"malware": 1, "vulnerability": 0, "phishing": -1,
		},
	}
}

func (cs *CriticalityScorer) Calculate(content string, categories []string) int {
	contentLower := strings.ToLower(content)
	
	maxScore := 0
	for keyword, keywordScore := range cs.keywords {
		if strings.Contains(contentLower, keyword) {
			if keywordScore > maxScore {
				maxScore = keywordScore
			}
		}
	}
	
	// default score if no keywords found
	if maxScore == 0 {
		maxScore = 3
	}
	
	// apply category modifiers
	for _, category := range categories {
		if modifier, exists := cs.categoryModifiers[category]; exists {
			maxScore += modifier
		}
	}
	
	// clamp to valid range
	if maxScore < 1 {
		maxScore = 1
	}
	if maxScore > 10 {
		maxScore = 10
	}
	
	return maxScore
}

func (cs *CriticalityScorer) AutoCategorize(content string) []string {
	contentLower := strings.ToLower(content)
	categories := []string{}

	// check for ransomware indicators
	if strings.Contains(contentLower, "ransomware") || 
	   strings.Contains(contentLower, "ransom") || 
	   strings.Contains(contentLower, "encrypt") {
		categories = append(categories, "ransomware")
	}

	// check for data leak indicators
	if strings.Contains(contentLower, "data breach") || 
	   strings.Contains(contentLower, "leak") || 
	   strings.Contains(contentLower, "stolen data") || 
	   strings.Contains(contentLower, "database dump") {
		categories = append(categories, "data-leak")
	}

	// check for malware indicators
	if strings.Contains(contentLower, "malware") || 
	   strings.Contains(contentLower, "trojan") || 
	   strings.Contains(contentLower, "virus") || 
	   strings.Contains(contentLower, "backdoor") {
		categories = append(categories, "malware")
	}

	// check for vulnerability indicators
	if strings.Contains(contentLower, "vulnerability") || 
	   strings.Contains(contentLower, "cve") || 
	   strings.Contains(contentLower, "zero-day") || 
	   strings.Contains(contentLower, "zero day") {
		categories = append(categories, "vulnerability")
	}

	// check for exploit indicators
	if strings.Contains(contentLower, "exploit") || 
	   strings.Contains(contentLower, "rce") || 
	   strings.Contains(contentLower, "remote code execution") {
		categories = append(categories, "exploit")
	}

	// check for phishing indicators
	if strings.Contains(contentLower, "phishing") || 
	   strings.Contains(contentLower, "phish") || 
	   strings.Contains(contentLower, "social engineering") {
		categories = append(categories, "phishing")
	}

	// default category if nothing matched
	if len(categories) == 0 {
		categories = append(categories, "vulnerability")
	}

	return categories
}
