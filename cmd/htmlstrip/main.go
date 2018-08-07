package main

import (
	"errors"
	"io"
	"os"

	"github.com/millerlogic/htmlstrip"
)

func run() error {
	var input io.Reader
	{
		var inputs []io.Reader
		switches := true
		for _, arg := range os.Args[1:] {
			if switches && len(arg) > 1 && arg[0] == '-' {
				if arg == "--" {
					switches = false
				} else {
					return errors.New("Invalid switch: " + arg)
				}
			} else {
				if arg == "-" {
					inputs = append(inputs, os.Stdin)
				} else {
					f, err := os.Open(arg)
					if err != nil {
						return err
					}
					defer f.Close()
					inputs = append(inputs, f)
				}
			}
		}
		if inputs == nil {
			input = os.Stdin
		} else {
			input = io.MultiReader(inputs...)
		}
	}

	w := &htmlstrip.Writer{W: os.Stdout}
	_, err := io.Copy(w, input)
	return err
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
