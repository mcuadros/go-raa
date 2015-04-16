package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/dustin/go-humanize"
)

const writeFlagsDefault = os.O_WRONLY | os.O_CREATE | os.O_TRUNC | os.O_EXCL
const writeFlagsOverwrite = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
const defaultPerms = 0755

type CmdUnpack struct {
	cmd
	Verbose     bool   `short:"v" description:"Activates the verbose mode"`
	Overwrite   bool   `short:"o" description:"Overwrites the files if arleady exists"`
	IgnorePerms bool   `short:"i" description:"Ignore files permisisions"`
	Match       string `short:"m" description:"Only extract files matching the given regexp"`

	Output struct {
		Path string `positional-arg-name:"output" description:"files or directories to be add to the archive."`
	} `positional-args:"yes"`

	flags        int
	regexp       *regexp.Regexp
	matchingFunc func(string) bool
}

func (c *CmdUnpack) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.buildArchive(); err != nil {
		return err
	}

	if err := c.do(); err != nil {
		return err
	}

	return nil
}

func (c *CmdUnpack) validate() error {
	err := c.cmd.validate()
	if err != nil {
		return err
	}

	if _, err := os.Stat(c.Args.File); err != nil {
		return fmt.Errorf("Invalid input file %q, %s\n", c.Args.File, err.Error())
	}

	if c.Output.Path == "" {
		c.Output.Path = "."
	}

	c.flags = writeFlagsDefault
	if c.Overwrite {
		c.flags = writeFlagsOverwrite
	}

	c.matchingFunc = func(string) bool { return true }
	if c.Match != "" {
		c.regexp, err = regexp.Compile(c.Match)
		if err != nil {
			return fmt.Errorf("Invalid match regexp %q, %s\n", c.Match, err.Error())
		}

		c.matchingFunc = func(name string) bool {
			return c.regexp.MatchString(name)
		}
	}

	return nil
}

func (c *CmdUnpack) do() error {
	for _, fname := range c.a.Find(c.matchingFunc) {
		c.extract(fname)
	}

	return nil
}

func (c *CmdUnpack) extract(srcName string) {
	src, err := c.a.Open(srcName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %q for read: %s\n", srcName, err.Error())
		return
	}

	defer src.Close()

	fi, _ := src.Stat()

	dstName := filepath.Join(c.Output.Path, srcName)
	dir := filepath.Dir(dstName)
	if err = os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create dir %q: %s\n", dir, err.Error())
		return
	}

	perms := os.FileMode(defaultPerms)
	if !c.IgnorePerms {
		perms = fi.Mode().Perm()
	}

	dst, err := os.OpenFile(dstName, c.flags, perms)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open %q for writing: %s\n", dstName, err.Error())
		return
	}

	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write %q : %s\n", srcName, err.Error())
		return
	}

	if c.Verbose {
		fmt.Println(srcName, humanize.Bytes(uint64(fi.Size())))
	}
}
