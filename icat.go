// The icat command displays an image to the terminal using block characters.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"slices"

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
	m = flagPrintMode(flag.CommandLine, "m", modeBlock24, "one of 'block', 'block24', or 'sixel'")
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

func sign(n int) float64 {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}

func floydRivest[S ~[]E, E any](array S, left, right, k int, cmp func(E, E) int) {
	for right > left {
		if right-left > 600 {
			n := right - left + 1
			i := k - left + 1
			z := math.Log(float64(n))
			s := .5 * math.Exp(2*z/3)
			sd := .5 * math.Sqrt(z*s*(float64(n)-s)/float64(n)) * sign(i-n/2)
			newLeft := max(left, int(float64(k)-float64(i)*s/float64(n)+sd))
			newRight := min(right, int(float64(k)+float64(n-i)*s/float64(n)+sd))
			floydRivest(array, newLeft, newRight, k, cmp)
		}
		t := array[k]
		i := left
		j := right
		array[left], array[k] = array[k], array[left]
		if cmp(array[right], t) > 0 {
			array[right], array[left] = array[left], array[right]
		}
		for i < j {
			array[i], array[j] = array[j], array[i]
			i++
			j--
			for ; cmp(array[i], t) < 0; i++ {
			}
			for ; cmp(array[j], t) > 0; j-- {
			}
		}
		if cmp(array[left], t) == 0 {
			array[left], array[j] = array[j], array[left]
		} else {
			j++
			array[j], array[right] = array[right], array[j]
		}
		if j <= k {
			left = j + 1
		}
		if k <= j {
			right = j - 1
		}
	}
}

func quickSelect[S ~[]E, E any](list S, k int, cmp func(E, E) int) {
	floydRivest(list, 0, len(list)-1, k, cmp)
}

func bucketRange(colors []color.RGBA) color.RGBA {
	if len(colors) == 0 {
		return color.RGBA{}
	}
	var minR, minG, minB uint8 = math.MaxUint8, math.MaxUint8, math.MaxUint8
	var maxR, maxG, maxB uint8
	for _, c := range colors {
		minR, maxR = min(minR, c.R), max(maxR, c.R)
		minG, maxG = min(minG, c.G), max(maxG, c.G)
		minB, maxB = min(minB, c.B), max(maxB, c.B)
	}
	return color.RGBA{R: maxR - minR, G: maxG - minG, B: maxB - minB}
}

func cutOnce(colors []color.RGBA, bucketRange color.RGBA) [2][]color.RGBA {
	if len(colors) == 0 {
		return [...][]color.RGBA{colors, colors}
	}
	rRange, gRange, bRange := bucketRange.R, bucketRange.G, bucketRange.B
	if rRange >= gRange && rRange >= bRange {
		quickSelect(colors, len(colors)/2, func(x, y color.RGBA) int { return int(x.R) - int(y.R) })
	} else if gRange >= rRange && gRange >= bRange {
		quickSelect(colors, len(colors)/2, func(x, y color.RGBA) int { return int(x.G) - int(y.G) })
	} else {
		quickSelect(colors, len(colors)/2, func(x, y color.RGBA) int { return int(x.B) - int(y.B) })
	}
	return [...][]color.RGBA{colors[:len(colors)/2], colors[len(colors)/2:]}
}

func colorAvg(colors []color.RGBA) color.RGBA {
	var r, g, b int64
	for _, c := range colors {
		r += int64(c.R)
		g += int64(c.G)
		b += int64(c.B)
	}
	n := int64(len(colors))
	return color.RGBA{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n), A: 0xff}
}

func medianCut(img image.Image) color.Palette {
	var colors []color.RGBA
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a > 0 {
				colors = append(colors, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0xff})
			}
		}
	}
	buckets := [][]color.RGBA{colors}
	bucketRanges := []color.RGBA{{}}
	for {
		var bestRange uint8
		var bestIdx int
		for i, rng := range bucketRanges {
			r := max(rng.R, rng.G, rng.B)
			if r >= bestRange {
				bestRange = r
				bestIdx = i
			}
		}
		split := cutOnce(buckets[bestIdx], bucketRanges[bestIdx])
		buckets = slices.Replace(buckets, bestIdx, bestIdx+1, split[:]...)
		if len(buckets) == 255 {
			break
		}
		bucketRanges = slices.Replace(bucketRanges, bestIdx, bestIdx+1, bucketRange(split[0]), bucketRange(split[1]))
	}
	palette := color.Palette{color.Transparent}
	for _, b := range buckets {
		if len(b) > 0 {
			palette = append(palette, colorAvg(b))
		}
	}
	return palette
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
		sixel.Print(os.Stdout, dst, medianCut(dst))
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
		if cellX == 0 && cellY == 0 {
			cellX, cellY = 10, 20
		}
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
