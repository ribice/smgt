package rot

import "strconv"

func numToString(age int) string {
	var name string // want "variable name should be declared right before it is used"
	if age < 18 {
		return "minor"
	}
	name = strconv.Itoa(age)
	return name
}

func inline(age int) string {
	if age < 18 {
		return "minor"
	}

	name := strconv.Itoa(age)
	return name
}
