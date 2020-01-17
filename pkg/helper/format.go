package helper

import (
	"strings"

	"github.com/kennygrant/sanitize"
)

func sanitizeText(input string) string {
	output := sanitize.BaseName(input)
	output = strings.ReplaceAll(output, "-", " ")
	output = strings.ToLower(output)

	return output
}
