package consts

import "strings"

var devmode string = "true"

func IsDevMode() bool {
	return strings.ToLower(devmode) == "true"
}
