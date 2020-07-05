package rules

import (
	"bufio"
	"os"
	"strings"

	C "github.com/Dreamacro/clash/constant"
)

func NewDomainSuffixSet(filename string, adapter string) (rules []C.Rule, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r := DomainSuffix{
			suffix:  strings.ToLower(scanner.Text()),
			adapter: adapter,
		}
		rules = append(rules, &r)
	}

	if err = scanner.Err(); err != nil {
		return
	}
	return
}
