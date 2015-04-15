package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewNamedParser("raa", flags.Default)
	parser.AddCommand("pack", "Create a new archive containing the specified items.", "", &CmdPack{})
	parser.AddCommand("info", "Display stats and/or the items contained on a file.", "", &CmdInfo{})

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}
