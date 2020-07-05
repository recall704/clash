package rules

import (
	"bufio"
	"net"
	"os"

	C "github.com/Dreamacro/clash/constant"
)

func NewIPCIDRSet(filename string, adapter string, opts ...IPCIDROption) (rules []C.Rule, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, ipnet, err := net.ParseCIDR(scanner.Text())
		if err != nil {
			return nil, errPayload
		}

		ipcidr := IPCIDR{
			ipnet:   ipnet,
			adapter: adapter,
		}

		for _, o := range opts {
			o(&ipcidr)
		}

		rules = append(rules, &ipcidr)
	}

	if err = scanner.Err(); err != nil {
		return
	}
	return
}
