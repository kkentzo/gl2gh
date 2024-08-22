package main

import (
	"os"

	"github.com/kkentzo/gl-to-gh/cmd"
)

var mappings = map[int]string{
	482361:  "kkentzo",
	2369470: "tkaretsos",
	1487483: "stsakanikas",
}

func main() {
	root := cmd.New()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
