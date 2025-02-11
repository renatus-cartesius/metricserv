package main

import "os"

func main() {
	os.Exit(1) // want "using os.Exit in main function"
}
