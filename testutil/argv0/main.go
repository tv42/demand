package main

import (
	"debug/elf"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	arg0 := os.Args[0]
	fmt.Println(filepath.Base(arg0))

	// make sure we can open the binary; lots of asset bundling
	// solutions need this
	f, err := os.Open(arg0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open arg0: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()
	// read just enough to make sure it's not a demand script
	buf := make([]byte, 4)
	if _, err := io.ReadFull(f, buf); err != nil {
		fmt.Fprintf(os.Stderr, "cannot read arg0: %v\n", err)
		os.Exit(1)
	}
	// this is annoyingly platform-specific, but i'd rather test this
	// way than compare against known-bad; false positive is better
	// than false negative for bug indicators
	if string(buf) != elf.ELFMAG {
		fmt.Fprintf(os.Stderr, "arg0 is not the binary: not ELF but %q\n", buf[:])
		os.Exit(1)
	}
}
