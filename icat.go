// The icat command displays an image to the terminal using block characters.
package main

import (
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
	"golang.org/x/sys/unix"
	"roseh.moe/pkg/sixel"
)

var (
	x = flag.Int("x", 0, "set image width in columns")
	y = flag.Int("y", 0, "set image height in rows")
	m = flagPrintMode(flag.CommandLine, "m", modeSixel, "one of 'block', 'block24', or 'sixel'")
)

type printMode int

const (
	modeInvalid printMode = iota
	modeBlock
	modeBlock24
	modeSixel
)

type printModeValue printMode

func (m *printModeValue) String() string {
	switch printMode(*m) {
	case modeBlock:
		return "block"
	case modeBlock24:
		return "block24"
	case modeSixel:
		return "sixel"
	default:
		return "invalid"
	}
}

func (m *printModeValue) Set(s string) error {
	var mode printMode
	switch s {
	case "block":
		mode = modeBlock
	case "block24":
		mode = modeBlock24
	case "sixel":
		mode = modeSixel
	default:
		return fmt.Errorf("bad mode type %q, should be one of 'block', 'block24', or 'sixel'", s)
	}
	*m = printModeValue(mode)
	return nil
}

func flagPrintMode(fs *flag.FlagSet, name string, value printMode, usage string) *printMode {
	fs.Var((*printModeValue)(&value), name, usage)
	return &value
}

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

func printImg(img image.Image, x, y, pixelX, pixelY int) error {
	// Try not to stretch the image.
	if y == 0 || x != 0 && img.Bounds().Dy()*x <= img.Bounds().Dx()*y*pixelY/pixelX {
		y = img.Bounds().Dy() * x / (img.Bounds().Dx() * pixelY / pixelX)
	} else {
		x = img.Bounds().Dx() * y * pixelY / pixelX / img.Bounds().Dy()
	}

	dst := image.NewRGBA(image.Rect(0, 0, x, y))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)

	switch *m {
	case modeBlock:
		sixel.PrintXTerm16(os.Stdout, dst)
	case modeBlock24:
		sixel.PrintBlock(os.Stdout, dst)
	case modeSixel:
		sixel.Print(os.Stdout, dst)
	}
	fmt.Println()
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

	cols := *x
	lines := *y
	pixelX := 2
	pixelY := 5
	if cols == 0 && lines == 0 {
		ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
		if err != nil {
			return err
		}
		cols = int(ws.Col)
		lines = int(ws.Row)
		lines-- // Leave a line for the status bar.
		cellX, cellY := int(ws.Xpixel)/int(ws.Col), int(ws.Ypixel)/int(ws.Row)
		if *m == modeSixel {
			cols *= cellX
			lines *= cellY
		} else {
			pixelX, pixelY = cellX, cellY
		}
	}
	switch *m {
	case modeSixel:
		pixelX, pixelY = 1, 1
	case modeBlock24:
		lines *= 2
		pixelX *= 2
	}

	img, err := load(file)
	if err != nil {
		return err
	}

	if err := printImg(img, cols, lines, pixelX, pixelY); err != nil {
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
