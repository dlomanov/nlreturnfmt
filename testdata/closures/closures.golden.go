package main

import "fmt"

func withGoroutine() {
	msg := "hello"
	go func() {
		fmt.Println(msg)

		return // A blank line should be inserted here.
	}()
}

func withDefer() {
	defer func() {
		fmt.Println("cleaning up")

		return // And here.
	}()

	fmt.Println("doing work")
}
