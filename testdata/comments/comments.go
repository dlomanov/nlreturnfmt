package main

func withComment() int {
	x := 10
	// nlreturnfmt intentionally does not add a blank line here.
	// This mirrors the original nlreturn linter's behavior, which is not comment-aware.
	return x
}
