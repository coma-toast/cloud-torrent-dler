package helper

import (
	"strings"

	"github.com/kennygrant/sanitize"
)

// SanitizeText eliminates unnecessary formatting.
func SanitizeText(input string) string {
	output := sanitize.BaseName(input)
	output = strings.ToLower(input)
	output = replaceText(output)

	return output
}

// SanitizePath is like SanitizeText, but without the / replaced
func SanitizePath(input string) string {
	array := strings.SplitAfter(input, "/")
	array[len(array)-1] = SanitizeText(array[len(array)-1])
	output := strings.Join(array, "")

	return output
}

func replaceText(input string) string {
	output := strings.ReplaceAll(input, "-", " ")
	output = strings.ReplaceAll(output, "_", " ")
	output = strings.ReplaceAll(output, ".", " ")
	// For some reason, a lot of file names don't quite match the magnet name - WEBRip in the magnet name is changed to webdl in the file name. Why?
	output = strings.ReplaceAll(output, "webrip", "")
	output = strings.ReplaceAll(output, "web dl", "")
	// Sometimes it's H.264, sometimes it's H264. Why??
	output = strings.ReplaceAll(output, "h 264", "")
	output = strings.ReplaceAll(output, "h.264", "")
	output = strings.ReplaceAll(output, "h264", "")
	// Sometimes it's h264, sometimes it's x264. Why???
	output = strings.ReplaceAll(output, "x264", "")
	// [] causes crashes. yay.
	output = strings.ReplaceAll(output, "[", "")
	output = strings.ReplaceAll(output, "]", "")
	// " -" now becomes "  " which is not right. Make it single space
	output = strings.ReplaceAll(output, "  ", " ")
	// as a result of all this nonsense, sometimes there are multiple . in a row. Fix that.
	output = strings.ReplaceAll(output, "..", ".")

	return output
}
