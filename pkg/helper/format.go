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
	// For some reason, a lot of file names don't quite match the magnet name - WEBRip in the magnet name is changed to webdl in the file name. Why?
	output = strings.ReplaceAll(output, "webrip", "")
	output = strings.ReplaceAll(output, "web dl", "")
	// " -" now becomes "  " which is not right. Make it single space
	output = strings.ReplaceAll(output, "  ", " ")
	// Sometimes it's H.264, sometimes it's H264. Why??
	output = strings.ReplaceAll(output, "h 264", "h264")
	// Sometimes it's h264, sometimes it's x264. Why???
	output = strings.ReplaceAll(output, "x264", "h264")

	return output
}
