package main

import (
	"fmt"
	"strconv"
)

type FractionMod struct {
}

func (*FractionMod) Name() string {
	return "fraction"
}

func (*FractionMod) Apply(s string) (string, error) {
	var suffix string
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if '0' <= c && c <= '9' {
			suffix = s[i+1:]
			s = s[:i+1]
			break
		}
	}
	num, denom := split2(s, '/')
	if denom == "" {
		denom = "1"
	}
	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return "", fmt.Errorf("unexpected numerator: %q", num)
	}
	d, err := strconv.ParseInt(denom, 10, 64)
	if err != nil {
		return "", fmt.Errorf("unexpected denominator: %q", denom)
	}

	x := gcd(n, d)

	return fmt.Sprintf("%d/%d%s", n/x, d/x, suffix), nil
}

func gcd(a, b int64) int64 {
	for b > 0 {
		n := a % b
		a = b
		b = n
	}
	return a
}
