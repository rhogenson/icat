package main

import (
	"bufio"
	"fmt"
	"golang.org/x/image/draw"
	"golang.org/x/term"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

var stdout = bufio.NewWriter(os.Stdout)

func load(filename string) (image.Image, error) {
	file := os.Stdin
	if filename != "-" {
		var err error
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode %q: %s", filename, err)
	}

	return img, nil
}

func printImg(filename string, cols, lines int) error {
	img, err := load(filename)
	if err != nil {
		return err
	}

	// Try not to stretch the image.
	if img.Bounds().Dy()*cols <= img.Bounds().Dx()*lines*5/2 {
		lines = img.Bounds().Dy() * cols / (img.Bounds().Dx() * 5 / 2)
	} else {
		cols = img.Bounds().Dx() * lines * 5 / 2 / img.Bounds().Dy()
	}

	dst := image.NewRGBA(image.Rect(0, 0, cols, lines))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	for y := 0; y < lines; y++ {
		for x := 0; x < cols; x++ {
			r, g, b, _ := dst.At(x, y).RGBA()
			fmt.Fprintf(stdout, "\033[48;2;%d;%d;%dm ", r>>8, g>>8, b>>8)
		}
		fmt.Fprintln(stdout, "\033[49m")
	}
	return nil
}

func icat(args []string) error {
	cols, lines, err := term.GetSize(1)
	if err != nil {
		return fmt.Errorf("terminal size: %s", err)
	}
	lines-- // Leave a line for the status bar.

	if len(args) == 0 {
		args = []string{"-"}
	}

	defer stdout.Flush()
	for _, f := range args {
		if err := printImg(f, cols, lines); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := icat(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
		os.Exit(1)
	}
}
