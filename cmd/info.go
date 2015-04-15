package main

import (
	"fmt"

	"github.com/mcuadros/raa"

	"github.com/dustin/go-humanize"
)

type CmdInfo struct {
	Stats bool `short:"s" long:"stats" description:"display stats about the file"`
	List  bool `short:"l" long:"list" description:"list the content stored in the file"`

	Args struct {
		File string `positional-arg-name:"output" required:"true" description:"a raa file."`
	} `positional-args:"yes"`

	v *raa.Volume
}

func (c *CmdInfo) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.buildVolume(); err != nil {
		return err
	}

	if err := c.listVolume(); err != nil {
		return err
	}

	return nil
}

func (c *CmdInfo) buildVolume() error {
	v, err := raa.NewVolume(c.Args.File)
	if err != nil {
		return err
	}

	c.v = v
	return nil
}

func (c *CmdInfo) listVolume() error {
	for _, file := range c.v.Find(func(string) bool { return true }) {
		fi, _ := c.v.Stat(file)
		fmt.Println(fi.Mode().Perm(), humanize.Bytes(uint64(fi.Size())), file)
	}

	return nil
}

func (c *CmdInfo) validate() error {
	if c.Args.File == "" {
		return fmt.Errorf("Invalid file %q", c.Args.File)
	}

	if !c.Stats && !c.List {
		c.Stats = true
	}

	return nil
}
