package rot

func elseIfWithInit(a int) int {
	if x := a; x > 10 {
		return x
	} else if y := a / 2; y > 5 {
		return y
	}
	return 0
}
