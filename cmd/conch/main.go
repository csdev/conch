package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/csdev/conch/internal/commit"
)

func main() {
	flag.Parse()
	commits, err := commit.ParseRange(flag.Arg(0))
	if err != nil {
		log.Fatalf("%v", err)
	}
	for _, c := range commits {
		fmt.Printf("%s: %s\n", c.Id[:7], c.Summary())
	}
}
