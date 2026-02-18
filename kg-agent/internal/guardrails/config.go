package guardrails

var DefaultBanWords = []string{
	// Profanity
	"fuck", "shit", "asshole",

	// Harmful instructions
	"hack", "exploit", "crack password",

	// Violence
	"kill", "murder", "bomb",

	// Add more...
}
