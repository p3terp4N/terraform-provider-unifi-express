package utils

import (
	"regexp"
	"strings"
)

var MacAddressRegexp = regexp.MustCompile("^([0-9a-fA-F][0-9a-fA-F][-:]){5}([0-9a-fA-F][0-9a-fA-F])$")

func CleanMAC(mac string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ToLower(mac), "-", ":"))
}
