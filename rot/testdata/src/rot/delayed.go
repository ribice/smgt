package rot

import (
	"strconv"
)

func guardWithReturn(age int) string {
	var label string // want "variable label should be declared right before it is used"
	if age < 0 {
		return "invalid"
	}
	helper()
	label = strconv.Itoa(age)
	return label
}

func loopBeforeUse(nums []int) string {
	var out string // want "variable out should be declared right before it is used"
	for _, n := range nums {
		if n < 0 {
			break
		}
	}
	out = strconv.Itoa(len(nums))
	return out
}

func switchBeforeUse(n int) string {
	var s string // want "variable s should be declared right before it is used"
	switch {
	case n < 0:
		return "negative"
	case n == 0:
		return "zero"
	}
	extraWork()
	s = strconv.Itoa(n)
	return s
}

func deferBeforeUse() bool {
	var ready bool // want "variable ready should be declared right before it is used"
	defer helper()
	ready = true
	return ready
}

func selectBeforeUse(ch <-chan struct{}) string {
	var status string // want "variable status should be declared right before it is used"
	select {
	case <-ch:
		helper()
	default:
		helper()
	}
	status = "done"
	return status
}

func goBeforeUse() int {
	var count int // want "variable count should be declared right before it is used"
	go func() {
		helper()
	}()
	count = 1
	return count
}

func shortDeclareLate(age int) string {
	name := strconv.Itoa(age) // want "variable name should be declared right before it is used"
	extraWork()
	return name
}

func varInitLate(age int) string {
	var label = strconv.Itoa(age) // want "variable label should be declared right before it is used"
	helper()
	return label
}

func funcLiteralDelayed() func() string {
	helloFunc := makeGreeter() // want "variable helloFunc should be declared right before it is used"
	extraWork()
	return helloFunc
}

func ifInitDelayed(input string) string {
	if value := makeValue(input); allowLoop() { // want "variable value should be declared right before it is used"
		helper()
		return value
	}
	return "empty"
}

func forInitDelayed(limit int) int {
	for index := 0; allowLoop(); index++ { // want "variable index should be declared right before it is used"
		helper()
		return index
	}
	return 0
}

func rangeDelayed(nums []int) int {
	for idx := range nums { // want "variable idx should be declared right before it is used"
		helper()
		return idx
	}
	return -1
}

func selectCaseDelayed(ch <-chan string) string {
	select {
	case msg := <-ch: // want "variable msg should be declared right before it is used"
		helper()
		return msg
	default:
		return "none"
	}
}

func typeSwitchDelayed(v any) string {
	switch val := v.(type) { // want "variable val should be declared right before it is used"
	case string:
		helper()
		return val
	default:
		return ""
	}
}

func helper() {}

func extraWork() {}

func makeGreeter() func() string {
	return func() string { return "hello" }
}

func makeValue(in string) string {
	return in
}

func allowLoop() bool {
	return true
}
