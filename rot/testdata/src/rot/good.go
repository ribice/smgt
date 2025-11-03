package rot

import (
	"strconv"
	"strings"
)

func immediateUse(age int) string {
	var label string
	label = strconv.Itoa(age)
	return label
}

func usedInsideBranches(age int) string {
	var label string
	if age < 0 {
		label = "invalid"
	} else {
		label = strconv.Itoa(age)
	}
	return label
}

func usedInSwitch(n int) string {
	var out string
	switch {
	case n < 0:
		out = "neg"
	default:
		out = strconv.Itoa(n)
	}
	return out
}

func usedInLoop(nums []int) int {
	var sum int
	for _, n := range nums {
		sum += n
	}
	return sum
}

func builderImmediateWork(parts []string) string {
	var builder strings.Builder
	builder.Grow(64)
	for _, p := range parts {
		builder.WriteString(p)
	}
	return builder.String()
}

func assignInsideBlock(flag bool) int {
	var value int
	{
		if flag {
			value = 2
		}
	}
	return value
}

func withLabel(flag bool) string {
	var out string
Label:
	if flag {
		out = "yes"
	} else {
		flag = true
		goto Label
	}
	return out
}

func usedInSelect(ch <-chan struct{}) string {
	var out string
	select {
	case <-ch:
		out = "done"
	default:
		out = "waiting"
	}
	return out
}

func shortDeclareImmediate(age int) string {
	name := strconv.Itoa(age)
	return name
}

func varInitImmediate(age int) string {
	var label = strconv.Itoa(age)
	return label
}

func ifInitImmediate(input string) string {
	if value := makeValue(input); value != "" {
		return value
	}
	return "empty"
}

func forInitImmediate(limit int) int {
	for index := 0; index < limit; index++ {
		return index
	}
	return 0
}

func rangeImmediate(nums []int) int {
	for idx := range nums {
		return idx
	}
	return -1
}

func selectCaseImmediate(ch <-chan string) string {
	select {
	case msg := <-ch:
		return msg
	default:
		return "none"
	}
}

func typeSwitchImmediate(v any) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return ""
	}
}
