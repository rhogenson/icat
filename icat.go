package main

import (
	"bufio"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/image/draw"
	"golang.org/x/term"
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

	dst := image.NewRGBA(image.Rect(0, 0, cols, 2*lines))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	for y := 0; y < 2*lines; y += 2 {
		for x := 0; x < cols; x++ {
			hiR, hiG, hiB, _ := dst.At(x, y).RGBA()
			loR, loG, loB, _ := dst.At(x, y+1).RGBA()
			fmt.Fprintf(stdout, "\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀",
				hiR>>8, hiG>>8, hiB>>8,
				loR>>8, loG>>8, loB>>8)
		}
		fmt.Fprintln(stdout, "\033[49m")
	}
	fmt.Fprint(stdout, "\033[39m")
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
