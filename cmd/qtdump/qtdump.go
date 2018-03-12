// qtdump.go
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"bitbucket.org/kleinnic74/photos/domain/formats"
)

var file string

func init() {
	flag.Parse()
	file = flag.Arg(0)
	if file == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Cannot open %s: %s", file, err)
	}
	defer f.Close()
	qt, err := formats.ReadAsQuicktime(f)
	if err != nil {
		panic(err)
	}
	var k = 0
	qt.Walk(func(a *formats.Atom, level int) {
		indent := strings.Repeat("  ", level)
		log.Printf("#%2d:%sAtom %s (%d bytes)", k, indent, a.TypeName(), a.SizeOfData())
		k += 1
	})
	log.Printf("DateTaken: %s", qt.DateTaken())
	log.Printf("Location: %s", qt.Location())
}
