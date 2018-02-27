package utils

import (
	"regexp"
	"strings"
)

func StripAndReverse(input string) string {
	//input = "https://myapp.pcf.io"
	hostname := splitOnDot(input)
	// hostname = "https://myapp"

	hostname = removeProtocol(hostname)
	// hostname = "//myapp"

	hostname = removeSpecialCharacters(hostname)
	// hostname = "myapp"

	hostname = reverseString(hostname)
	// hostname = "ppaym"

	return hostname
}

func splitOnDot(input string) (result string) {
	delimiter := "."

	results := strings.Split(input, delimiter)

	result = results[0]

	return result
}

func removeProtocol(input string) string {
	return strings.Split(input, ":")[1]
}

func removeSpecialCharacters(input string) string {
	re := regexp.MustCompile("\\W")

	return re.ReplaceAllString(input, "")
}

func reverseString(input string) string {
	var result string

	for _, character := range input {
		result = string(character) + result
	}

	return result
}
