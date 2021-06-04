package cert

var challenges map[string]string

func GetChallenge(domain string) string {
	return challenges[domain]
}

func SetChallenge(domain string, challenge string) {
	challenges[domain] = challenge
}
