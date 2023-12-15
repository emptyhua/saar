package main

import (
	"fmt"
	"os"

	"github.com/emptyhua/saar"
)

func usage() {
	fmt.Fprintf(os.Stderr, "saar c <dst.saar> <file1> [file2] ...\n")
	fmt.Fprintf(os.Stderr, "saar x <src.saar> <dst dir>\n")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 4 {
		usage()
	}

	var err error
	if os.Args[1] == "c" {
		progFunc := func(p1 string, p2 string, err error) {
			fmt.Printf("add %s %s %v\n", p1, p2, err)
		}

		err = saar.CreateArchive(progFunc, os.Args[2], os.Args[2:]...)
	} else if os.Args[1] == "x" {
		progFunc := func(p1 string, p2 string, err error) {
			fmt.Printf("extract %s %s %v\n", p1, p2, err)
		}

		err = saar.ExtractArchive(progFunc, os.Args[2], os.Args[3])
	} else {
		usage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
