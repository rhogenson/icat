package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
	"golang.org/x/term"
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

func printImg(img image.Image, cols, lines int) error {
	// Try not to stretch the image.
	if img.Bounds().Dy()*cols <= img.Bounds().Dx()*lines*5/2 {
		lines = img.Bounds().Dy() * cols / (img.Bounds().Dx() * 5 / 2)
	} else {
		cols = img.Bounds().Dx() * lines * 5 / 2 / img.Bounds().Dy()
	}

	dst := image.NewRGBA(image.Rect(0, 0, cols, 2*lines))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for y := 0; y < 2*lines; y += 2 {
		for x := 0; x < cols; x++ {
			hiR, hiG, hiB, _ := dst.At(x, y).RGBA()
			loR, loG, loB, _ := dst.At(x, y+1).RGBA()
			fmt.Fprintf(w, "\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀",
				hiR>>8, hiG>>8, hiB>>8,
				loR>>8, loG>>8, loB>>8)
		}
		fmt.Fprintln(w, "\033[49m")
	}
	fmt.Fprint(w, "\033[39m")
	return nil
}

func icat(args []string) error {
	if len(args) == 0 {
		return errors.New("missing positional argument")
	}
	if len(args) > 1 {
		return errors.New("too many positional arguments")
	}
	file := args[0]

	cols, lines, err := term.GetSize(1)
	if err != nil {
		return fmt.Errorf("terminal size: %s", err)
	}
	lines-- // Leave a line for the status bar.

	img, err := load(file)
	if err != nil {
		return err
	}

	if err := printImg(img, cols, lines); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: icat FILE\n")
	}
	flag.Parse()

	if err := icat(flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
		os.Exit(1)
	}
}
