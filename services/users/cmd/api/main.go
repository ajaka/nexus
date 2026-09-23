package main

import "users/internal"

func main() {
	if err := internal.Listen(); err != nil {
		panic(err)
	}
}
