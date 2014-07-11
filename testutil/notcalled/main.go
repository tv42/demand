package main

import "os"

func main() {
	_, _ = os.Stdout.Write([]byte("never call me!\n"))
	os.Exit(42)
}
