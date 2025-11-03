package loopnow

import (
	t "time"
)

func positiveBody() {
	for i := 0; i < 3; i++ {
		now := t.Now() // want "time.Now should not be called inside loops; compute the value outside the loop"
		_ = now
	}
}

func positiveRange(xs []int) {
	for _, x := range xs {
		_ = x
		_ = t.Now().Unix() // want "time.Now should not be called inside loops; compute the value outside the loop"
	}
}

func positiveCondition(limit t.Duration) {
	start := t.Now()
	for t.Now().Before(start.Add(limit)) { // want "time.Now should not be called inside loops; compute the value outside the loop"
		return
	}
}

func positiveNested(limit int) {
	for i := 0; i < limit; i++ {
		func() {
			_ = t.Now() // want "time.Now should not be called inside loops; compute the value outside the loop"
		}()
	}
}

func negativeOutsideLoop() {
	start := t.Now()
	deadline := start.Add(10 * t.Second)
	for start.Before(deadline) {
		break
	}
}

func negativeDifferentPackage(now func() t.Time) {
	for i := 0; i < 5; i++ {
		_ = now()
	}
}

func negativeOtherCall(xs []int) {
	start := t.Now()
	for range xs {
		_ = t.Since(start)
	}
}
