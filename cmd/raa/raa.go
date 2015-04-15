package main

import (
	"fmt"
	"os"

	"github.com/mcuadros/raa"

	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewNamedParser("raa", flags.Default)
	parser.AddCommand("pack", "Create a new archive containing the specified items.", "", &CmdPack{})
	parser.AddCommand("list", "List the items contained on a file.", "", &CmdList{})
	parser.AddCommand("stats", "Display some stats about the file.", "", &CmdStats{})

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}

type cmd struct {
	Args struct {
		File string `positional-arg-name:"output" required:"true" description:"a raa file."`
	} `positional-args:"yes"`

	v *raa.Volume
}

func (c *cmd) validate() error {
	if c.Args.File == "" {
		return fmt.Errorf("Invalid raa file %q", c.Args.File)
	}
	return nil
}

func (c *cmd) buildVolume() error {
	v, err := raa.NewVolume(c.Args.File)
	if err != nil {
		return err
	}

	c.v = v
	return nil
}
