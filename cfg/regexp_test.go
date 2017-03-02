package cfg

import (
	"testing"
	"regexp"
)

func Test(t *testing.T) {
	regex,err := regexp.Compile(`app:(\d+):firstname`)
	if err != nil {
	}
	println(regex.Match([]byte("app:1223232:lastname")))
	println(regex.Match([]byte("app:1223232:firstname")))
	println(string(regex.FindSubmatch([]byte("app:1223232:firstname"))[1]))
	println(regex.FindSubmatch([]byte("app:1223232:lastname")))
}
