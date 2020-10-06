package helper

import (
	"strings"

	"github.com/kennygrant/sanitize"
)

// SanitizeText eliminates unnecessary formatting.
func SanitizeText(input string) string {
	output := sanitize.BaseName(input)
	output = strings.ReplaceAll(output, "-", " ")
	output = strings.ToLower(output)
	// For some reason, a lot of file names don't quite match the magnet name - WEBRip in the magnet name is changed to webdl in the file name.
	output = strings.ReplaceAll(output, "webrip", "")
	output = strings.ReplaceAll(output, "web dl", "")
	output = strings.ReplaceAll(output, "  ", " ")

	return output
}
