package helper

import "regexp"

var numRE = regexp.MustCompile("^[0-9]*(.[0-9])?$")

func IsValidNumber(s string) bool {
	return numRE.Match([]byte(s))
}
