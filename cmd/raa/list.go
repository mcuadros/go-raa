package main

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type CmdList struct {
	cmd
}

func (c *CmdList) Execute(args []string) error {
	if err := c.buildArchive(); err != nil {
		return err
	}

	defer c.a.Close()
	if err := c.listVolume(); err != nil {
		return err
	}

	return nil
}

func (c *CmdList) listVolume() error {
	for _, file := range c.a.Find(func(string) bool { return true }) {
		fi, _ := c.a.Stat(file)

		fmt.Printf("%s %s % 6s %s\n",
			fi.Mode().Perm(),
			fi.ModTime().Format("Jan 2 15:04"),
			humanize.Bytes(uint64(fi.Size())),
			file,
		)
	}

	return nil
}
