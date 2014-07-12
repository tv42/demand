package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if _, err := fmt.Println("ok"); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
