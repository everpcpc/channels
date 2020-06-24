package slack

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	reTest = regexp.MustCompile(`<(@|#)([0-9a-zA-Z]+)>`)
)

func reverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return
}

func TestReplaceAllStringSubmatchFunc(t *testing.T) {
	testString := "abc<@ABC>abc<#ijk>"
	result :=
		replaceAllStringSubmatchFunc(reTest, testString, func(groups []string) string {
			indicator := groups[1]
			name := groups[2]
			return fmt.Sprintf("<%s%s>", indicator, reverse(name))
		})
	assert.Equal(t, result, "abc<@CBA>abc<#kji>")

}
