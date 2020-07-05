package rules

import (
	"bufio"
	"os"
	"strconv"

	C "github.com/Dreamacro/clash/constant"
)

func NewPortSet(filename string, adapter string, isSource bool) (rules []C.Rule, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, err := strconv.Atoi(scanner.Text())
		if err != nil {
			return nil, errPayload
		}
		r := Port{
			adapter:  adapter,
			port:     scanner.Text(),
			isSource: isSource,
		}

		rules = append(rules, &r)
	}

	if err = scanner.Err(); err != nil {
		return
	}
	return
}
