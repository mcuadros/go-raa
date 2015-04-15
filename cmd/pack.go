package main

import (
	"fmt"
	"os"

	"github.com/mcuadros/raa"
)

type CmdPack struct {
	Args struct {
		Output string   `positional-arg-name:"output" description:"write the archive to the specified file."`
		Input  []string `positional-arg-name:"input" description:"files or directories to be add to the archive."`
	} `positional-args:"yes"`

	v *raa.Volume
}

func (c *CmdPack) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.do(); err != nil {
		if err := os.Remove(c.Args.Output); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (c *CmdPack) do() error {
	if err := c.buildVolume(); err != nil {
		return err
	}

	if err := c.processInputToVolume(); err != nil {
		return err
	}

	return nil
}

func (c *CmdPack) validate() error {
	if c.Args.Output == "" {
		return fmt.Errorf("Invalid output file %q", c.Args.Output)
	}

	if _, err := os.Stat(c.Args.Output); err == nil {
		return fmt.Errorf("Invalid output file %q, file already exists", c.Args.Output)
	}

	if len(c.Args.Input) == 0 {
		return fmt.Errorf("Invalid input count, please add one or more input files/dirs")
	}

	return nil
}

func (c *CmdPack) buildVolume() error {
	v, err := raa.NewVolume(c.Args.Output)
	if err != nil {
		return err
	}

	c.v = v
	return nil
}

func (c *CmdPack) processInputToVolume() error {
	target := "/"
	for _, file := range c.Args.Input {
		fi, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("Invalid input file/dir %q, no such file", file)
		}

		switch {
		case fi.Mode().IsRegular():
			_, err = raa.AddFile(c.v, file, target)
		case fi.Mode().IsDir():
			_, err = raa.AddDirectory(c.v, file, target, true)
		default:
			_, err = raa.AddGlob(c.v, file, target, true)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
