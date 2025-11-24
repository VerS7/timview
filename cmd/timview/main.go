package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"timview/internal/platform"
	"timview/internal/resize"
	"timview/internal/util"
)

const (
	ESCAPE  = "\033"
	SEGMENT = "▀"
	FG      = ESCAPE + "[38;2;%d;%d;%dm"
	BG      = ESCAPE + "[48;2;%d;%d;%dm"
	RESET   = ESCAPE + "[0m"
	CLEAR   = ESCAPE + "[H" + ESCAPE + "[2J"
	NEWLINE = "\n"

	HELP_TEXT = "" +
		`  _____ ___ __  ____   ___            ` + "\n" +
		` |_   _|_ _|  \/  \ \ / (_)_____ __ __` + "\n" +
		`   | |  | || |\/| |\ V /| / -_) V  V /` + "\n" +
		`   |_| |___|_|  |_| \_/ |_\___|\_/\_/ ` + "\n\n" +
		"Terminal\nIMage\nVIEWer\n\n" +
		"small program for viewing images (png, jpg) in terminal by using ANSI escape symbols\n" +
		"usage: timview [-r] [image]\n" +
		"params:\n"
)

var (
	ratio   = flag.Float64("r", 0.5, "aspect ratio of output image. Default is half of terminal width. Min = 0.1, Max = 1")
	samples = flag.Int("s", 2, "samples count for smoothing output image. Min = 2, Max = 16")
	nowarn  = flag.Bool("nowarn", false, "disables WARN messages")
)

func RenderImage(img image.Image) string {
	max := img.Bounds().Max

	var finalImg strings.Builder
	finalImg.Grow(max.Y / 2)

	var row strings.Builder
	row.Grow(max.X * len(FG+BG+SEGMENT))

	for y := 0; y < max.Y; y += 2 {
		row.Reset()

		for x := 0; x < max.X; x += 1 {
			ur, ug, ub, _ := img.At(x, y).RGBA()
			lr, lg, lb, _ := img.At(x, y+1).RGBA()

			var fragment string

			if y+1 >= max.Y {
				fragment = fmt.Sprintf(
					FG+SEGMENT,
					// Normalize to 0-255
					ur>>8,
					ug>>8,
					ub>>8,
				)
			} else {
				fragment = fmt.Sprintf(
					FG+BG+SEGMENT,
					// Normalize to 0-255
					ur>>8,
					ug>>8,
					ub>>8,
					lr>>8,
					lg>>8,
					lb>>8,
				)
			}
			row.WriteString(fragment)
		}
		finalImg.WriteString(row.String())
		finalImg.WriteString(RESET)
		finalImg.WriteString(NEWLINE)
	}

	finalImg.WriteString(RESET)
	return finalImg.String()
}

func ProcessImage(imagepath string, targetWidth int, scale float64) (image.Image, error) {
	// Open image file
	file, err := os.Open(imagepath)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s", imagepath)
	}
	defer file.Close()

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("could not decode image with format %s", format)
	}

	scale = util.Clamp(scale, 0.1, 1.0)

	newWidth := math.Round(float64(targetWidth) * scale)

	smpls := *samples
	smpls = util.Clamp(smpls, 2, 16)

	return resize.Resize(uint(newWidth), 0, img, resize.MitchellnetravaliInterp, smpls), nil
}

func main() {
	out := os.Stdout
	outFd := int(out.Fd())

	flag.Usage = func() {
		fmt.Fprint(out, HELP_TEXT)
		flag.PrintDefaults()
	}

	flag.Parse()

	disableWarnings := *nowarn

	// Enable colored output (platform specific)
	if colorEnabled := platform.EnableColoredOutput(outFd); !colorEnabled {
		panic("ERROR: colored output not supported")
	}

	// Check if FD is terminal
	if !platform.IsTerminal(outFd) {
		if !disableWarnings {
			fmt.Fprintln(out, "WARN: not in terminal")
		}
	}

	// Get terminal width for image scaling
	width, _, err := platform.GetSize(outFd)
	if err != nil {
		if !disableWarnings {
			fmt.Fprintln(out, "WARN: could not get terminal bounds. Width set to 100 symbols")
		}
		width = 100
	}

	imagepath := flag.Arg(0)
	if imagepath == "" {
		panic("ERROR: image path is empty")
	}

	img, err := ProcessImage(imagepath, width, *ratio)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}

	fmt.Fprint(out, CLEAR)
	fmt.Fprint(out, RenderImage(img))
}
