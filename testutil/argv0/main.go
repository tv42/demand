package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	p := filepath.Base(os.Args[0])
	fmt.Println(p)
}
