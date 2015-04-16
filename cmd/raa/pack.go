package main

import (
	"fmt"
	"os"

	"github.com/mcuadros/go-raa"
)

type CmdPack struct {
	cmd
	Input struct {
		Files []string `positional-arg-name:"input" description:"files or directories to be add to the archive."`
	} `positional-args:"yes"`
}

func (c *CmdPack) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.do(); err != nil {
		if err := os.Remove(c.Args.File); err != nil {
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
	if err := c.cmd.validate(); err != nil {
		return err
	}

	if _, err := os.Stat(c.Args.File); err == nil {
		return fmt.Errorf("Invalid output file %q, file already exists", c.Args.File)
	}

	if len(c.Input.Files) == 0 {
		return fmt.Errorf("Invalid input count, please add one or more input files/dirs")
	}

	return nil
}

func (c *CmdPack) processInputToVolume() error {
	target := "/"
	for _, file := range c.Input.Files {
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
