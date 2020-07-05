package rules

import (
	"fmt"

	C "github.com/Dreamacro/clash/constant"
)

func ParseRule(tp, payload, target string, params []string) ([]C.Rule, error) {
	var (
		parseErr error
		parsed   C.Rule
		rules    []C.Rule
	)

	switch tp {
	case "DOMAIN":
		parsed = NewDomain(payload, target)
		rules = append(rules, parsed)
	case "DOMAIN-SET":
		parsedRules, err := NewDomainSet(payload, target)
		if err == nil {
			rules = append(rules, parsedRules...)
		}
	case "DOMAIN-SUFFIX":
		parsed = NewDomainSuffix(payload, target)
		rules = append(rules, parsed)
	case "DOMAIN-SUFFIX-SET":
		parsedRules, parseErr := NewDomainSuffixSet(payload, target)
		if parseErr == nil {
			rules = append(rules, parsedRules...)
		}
	case "DOMAIN-KEYWORD":
		parsed = NewDomainKeyword(payload, target)
		rules = append(rules, parsed)
	case "DOMAIN-KEYWORD-SET":
		parsedRules, parseErr := NewDomainKeywordSet(payload, target)
		if parseErr == nil {
			rules = append(rules, parsedRules...)
		}
	case "GEOIP":
		noResolve := HasNoResolve(params)
		parsed = NewGEOIP(payload, target, noResolve)
		rules = append(rules, parsed)
	case "IP-CIDR", "IP-CIDR6":
		noResolve := HasNoResolve(params)
		parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRNoResolve(noResolve))
		if parseErr == nil {
			rules = append(rules, parsed)
		}
	case "IP-CIDR-SET", "IP-CIDR6-SET":
		noResolve := HasNoResolve(params)
		parsedRules, parseErr := NewIPCIDRSet(payload, target, WithIPCIDRNoResolve(noResolve))
		if parseErr == nil {
			rules = append(rules, parsedRules...)
		}
	case "SRC-IP-CIDR":
		parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRSourceIP(true), WithIPCIDRNoResolve(true))
		if parseErr == nil {
			rules = append(rules, parsed)
		}
	case "SRC-IP-CIDR-SET":
		parsedRules, parseErr := NewIPCIDRSet(payload, target, WithIPCIDRSourceIP(true), WithIPCIDRNoResolve(true))
		if parseErr == nil {
			rules = append(rules, parsedRules...)
		}
	case "SRC-PORT":
		parsed, parseErr = NewPort(payload, target, true)
		if parseErr == nil {
			rules = append(rules, parsed)
		}
	case "SRC-PORT-SET":
		parsedRules, parseErr := NewPortSet(payload, target, true)
		if parseErr == nil {
			rules = append(rules, parsedRules...)
		}
	case "DST-PORT":
		parsed, parseErr = NewPort(payload, target, false)
		if parseErr == nil {
			rules = append(rules, parsed)
		}
	case "DST-PORT-SET":
		parsedRules, parseErr := NewPortSet(payload, target, false)
		if parseErr == nil {
			rules = append(rules, parsedRules...)
		}
	case "MATCH":
		parsed = NewMatch(target)
		rules = append(rules, parsed)
	default:
		parseErr = fmt.Errorf("unsupported rule type %s", tp)
	}

	return rules, parseErr
}
