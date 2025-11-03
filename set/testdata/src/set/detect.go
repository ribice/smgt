package set

const alwaysTrue = true

func literalOnlyTrue() {
	m := map[string]bool{ // want "map\\[string\\]bool variable m is used as a set; use map\\[string\\]struct\\{\\} instead"
		"foo": true,
		"bar": true,
	}
	_ = m
}

func literalMixedValues() {
	m := map[string]bool{
		"foo": true,
		"bar": false,
	}
	_ = m
}

func literalNonConstant() {
	m := map[string]bool{
		"foo": compute(),
	}
	_ = m
}

func assignmentOnlyTrue(keys []string) {
	s := map[string]bool{} // want "map\\[string\\]bool variable s is used as a set; use map\\[string\\]struct\\{\\} instead"
	s["alpha"] = true
	for _, k := range keys {
		s[k] = alwaysTrue
	}
}

func assignmentWithFalse(keys []string) {
	s := map[string]bool{}
	s["alpha"] = true
	for _, k := range keys {
		s[k] = k == "beta"
	}
}

var packageLevel = map[string]bool{ // want "map\\[string\\]bool variable packageLevel is used as a set; use map\\[string\\]struct\\{\\} instead"
	"alpha": true,
}

func compute() bool {
	return len(packageLevel) > 0
}
