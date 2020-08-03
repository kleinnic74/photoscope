package consts

import "strings"

var devmode string = "false"

func IsDevMode() bool {
	return strings.ToLower(devmode) == "true"
}
