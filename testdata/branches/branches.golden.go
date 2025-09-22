package main

func withContinue() {
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			println(i)

			continue // A blank line should be inserted here.
		}
	}
}

func withGoto() {
	i := 0
Loop:
	if i < 5 {
		println(i)
		i++

		goto Loop // A blank line should be inserted here.
	}
}
