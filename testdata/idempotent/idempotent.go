package main

func alreadyCorrect() int {
	x := 1
	x++

	// A blank line already exists, so no changes should be made.
	return x
}

func firstStatement() int {
	// 'return' is the first statement in the block; no blank line needed.
	return 1
}

func onlyStatement() {
	// 'return' is the only statement in the block; no blank line needed.
	return
}

func shortBlock(blockSize int) string {
	// This block is shorter than the configured blockSize,
	// so no blank line is needed.
	// This test should be run with blockSize >= 2.
	s := "hello"
	return s
}
