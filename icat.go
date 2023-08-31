package main

import (
	"fmt"
	"golang.org/x/term"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

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

func printImg(filename string) error {
	img, err := load(filename)
	if err != nil {
		return err
	}

	cols, lines, err := term.GetSize(1)
	if err != nil {
		return fmt.Errorf("terminal size: %s", err)
	}
	lines-- // Leave a line for the status bar.

	// Try not to stretch the image.
	if img.Bounds().Dy()*cols <= img.Bounds().Dx()*lines*5/2 {
		lines = img.Bounds().Dy() * cols / (img.Bounds().Dx() * 5 / 2)
	} else {
		cols = img.Bounds().Dx() * lines * 5 / 2 / img.Bounds().Dy()
	}

	// nearest-neighbor interpolation
	for y := 0; y < lines; y++ {
		for x := 0; x < cols; x++ {
			sx := x*img.Bounds().Dx()/cols + img.Bounds().Min.X
			sy := y*img.Bounds().Dy()/lines + img.Bounds().Min.Y
			r, g, b, _ := img.At(sx, sy).RGBA()
			fmt.Printf("\033[48;2;%d;%d;%dm ", r>>8, g>>8, b>>8)
		}
		fmt.Println("\033[49m")
	}
	return nil
}

func icat(args []string) error {
	if len(args) == 0 {
		args = []string{"-"}
	}
	for _, f := range args {
		if err := printImg(f); err != nil {
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
