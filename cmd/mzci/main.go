package main

import (
	"fmt"
	"os"

	"github.com/katsyoshi/manazashi"
)

func main() {
	if err := manazashi.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
