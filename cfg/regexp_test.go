package cfg

import (
	"regexp"
	"testing"
)

func Test(t *testing.T) {
	regex, err := regexp.Compile(`app:(\d+):firstname`)
	if err != nil {
	}
	println(regex.MatchString("app:1223232:lastname"))
	println(regex.MatchString("app:1223232:firstname"))
	println(string(regex.FindStringSubmatch("app:1223232:firstname")[1]))
	println(regex.FindStringSubmatch("app:1223232:lastname"))
}
