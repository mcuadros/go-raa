package main

import (
	"fmt"
	"os"

	"github.com/mcuadros/go-raa"

	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewNamedParser("raa", flags.Default)
	parser.AddCommand("pack", "Create a new archive containing the specified items.", "", &CmdPack{})
	parser.AddCommand("unpack", "Extract to disk from the archive.", "", &CmdUnpack{})
	parser.AddCommand("list", "List the items contained on a file.", "", &CmdList{})
	parser.AddCommand("stats", "Display some stats about the file.", "", &CmdStats{})

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}

type cmd struct {
	Args struct {
		File string `positional-arg-name:"raa-file" required:"true" description:"raa file."`
	} `positional-args:"yes"`

	a *raa.Archive
}

func (c *cmd) validate() error {
	if c.Args.File == "" {
		return fmt.Errorf("Missing raa file, please provide a valid one.")
	}
	return nil
}

func (c *cmd) buildArchive() error {
	a, err := raa.CreateArchive(c.Args.File)
	if err != nil {
		return err
	}

	c.a = a
	return nil
}
