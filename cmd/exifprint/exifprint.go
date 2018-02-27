package main

import (
	"flag"
	"fmt"

	"bitbucket.org/kleinnic74/photos/domain"
)

func printTag(name, value string) {
	fmt.Printf("  %s=%s\n", name, value)
}

func main() {
	flag.Parse()
	fmt.Printf("%s:\n", flag.Arg(0))
	domain.PrintExif(flag.Arg(0), printTag)
}
