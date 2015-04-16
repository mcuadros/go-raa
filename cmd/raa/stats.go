package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

type CmdStats struct {
	cmd
}

func (c *CmdStats) Execute(args []string) error {
	if err := c.buildArchive(); err != nil {
		return err
	}

	defer c.a.Close()
	if err := c.displayStats(); err != nil {
		return err
	}

	return nil
}

func (c *CmdStats) displayStats() error {
	size, count := c.collectStats()

	tSize := sumMapInt64(size)
	tCount := sumMapInt64(count)

	fmt.Println("File:\t\t\t", c.a.Path())

	fi, err := os.Stat(c.a.Path())
	if err != nil {
		return err
	}

	fmt.Println("Number of files:\t", tCount)
	fmt.Println("Content size:\t\t", humanize.Bytes(uint64(tSize)))
	fmt.Println("RAA size:\t\t", humanize.Bytes(uint64(fi.Size())))

	ratio := float64(fi.Size()) / float64(tSize)
	fmt.Printf("Space saving ratio:\t %.2f%%\n\n", (1-ratio)*100)

	fmt.Printf("%10s %5s %6s %7s\n", "extension", "files", "size", "size %")
	fmt.Printf("-------------------------------\n")
	for ext, s := range size {
		fmt.Printf("%10s %5d %6s %6.2f%%\n",
			ext,
			count[ext],
			humanize.Bytes(uint64(s)),
			float64(s)/float64(tSize)*100,
		)
	}

	return nil
}

func (c *CmdStats) collectStats() (map[string]int64, map[string]int64) {
	size := make(map[string]int64, 0)
	count := make(map[string]int64, 0)

	for _, file := range c.a.Find(func(string) bool { return true }) {
		fi, _ := c.a.Stat(file)

		ext := filepath.Ext(file)
		if _, ok := size[ext]; !ok {
			size[ext] = 0
		}

		size[ext] += fi.Size()
		if _, ok := count[ext]; !ok {
			count[ext] = 0
		}

		count[ext]++
	}

	return size, count
}

func sumMapInt64(i map[string]int64) int64 {
	var t int64
	for _, v := range i {
		t += v
	}

	return t
}
