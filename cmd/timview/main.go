package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"timview/internal/platform"
	"timview/internal/resize"
	"timview/internal/util"
)

const HELP_TEXT = `
  _____ ___ __  ____   ___            
 |_   _|_ _|  \/  \ \ / (_)_____ __ __
   | |  | || |\/| |\ V /| / -_) V  V /
   |_| |___|_|  |_| \_/ |_\___|\_/\_/ 

Terminal
IMage
VIEWer

small program for viewing images (png, jpg) in terminal by using ANSI escape symbols
usage: timview [-r] [image]
params:
`

const (
	ESCAPE      = "\033"
	SEGMENT     = "▀"
	NEWLINE     = "\n"
	RETURN      = "\r"
	FG          = ESCAPE + "[38;2;%d;%d;%dm"
	BG          = ESCAPE + "[48;2;%d;%d;%dm"
	RESET       = ESCAPE + "[0m"
	CLEAR       = ESCAPE + "[2J" + ESCAPE + "[3J" + ESCAPE + "[H"
	CURSOR_HIDE = ESCAPE + "[?25l"
	CURSOR_SHOW = ESCAPE + "[?25h"
)

type KEY uint

const (
	LEFT KEY = iota
	RIGHT
	EXIT
)

var (
	ratio   = flag.Float64("r", 0.5, "aspect ratio of output image. Default is half of terminal width. Min = 0.1, Max = 1")
	samples = flag.Int("s", 2, "samples count for smoothing output image. Min = 2, Max = 16")
	width   = flag.Int("w", 0, "width of image in terminal symbols. Default is auto-detected, you rarely should specify it")
	nowarn  = flag.Bool("nowarn", false, "disables WARN messages")
	extra   = flag.Bool("extra", false, "extra image info")
	freeze  = flag.Bool("freeze", false, "freeze and wait for key after displaying image")
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var LevelNames = map[Level]string{DEBUG: "DEBUG", INFO: "INFO", WARN: "WARN", ERROR: "ERROR"}

type SymbolicImage struct {
	Path          string
	Repr          string
	Size          int64
	InitialBounds image.Rectangle
	FinalBounds   image.Rectangle
}

func (si *SymbolicImage) FormatExtra() string {
	return fmt.Sprintf(
		"Initial bounds: %dx%d"+NEWLINE+RETURN+"Final bounds: %dx%d"+NEWLINE+RETURN+"Size: %d bytes"+NEWLINE+RETURN,
		si.InitialBounds.Size().X,
		si.InitialBounds.Size().Y,
		si.FinalBounds.Size().X,
		si.FinalBounds.Size().Y,
		si.Size)
}

// Auto panics on ERROR level!
func Log(level Level, data ...any) {
	// Skip WARN level if --nowarn
	if *nowarn && level == WARN {
		return
	}

	message := fmt.Sprintf("[%s] ", LevelNames[level])
	for _, d := range data {
		message += fmt.Sprintf("%+v", d)
	}
	message += NEWLINE + RETURN

	switch level {
	case WARN:
		if *nowarn {
			return
		}
		fmt.Print(message)
	case INFO, DEBUG:
		fmt.Print(message)
	case ERROR:
		panic(message)
	}
}

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
		finalImg.WriteString(RETURN)
	}

	return finalImg.String()
}

func DecodeImage(imagepath string) (image.Image, int64, error) {
	// Open image file
	file, err := os.Open(imagepath)
	if err != nil {
		return nil, 0, fmt.Errorf("could not open file %s", imagepath)
	}
	defer file.Close()

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, 0, fmt.Errorf("could not decode image with format %s", format)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("could not get file stat of file %s", imagepath)
	}

	return img, stat.Size(), nil
}

func ScaleImage(img image.Image, targetWidth int, scale float64) (image.Image, error) {
	scale = util.Clamp(scale, 0.1, 1.0)

	newWidth := math.Round(float64(targetWidth) * scale)

	smpls := *samples
	smpls = util.Clamp(smpls, 2, 16)

	return resize.Resize(uint(newWidth), 0, img, resize.MitchellnetravaliInterp, smpls), nil
}

func PrintControls(out *os.File, currElem int, maxElems int, targetWidth int) {
	text := fmt.Sprintf("< %d/%d >", currElem, maxElems)
	tooltip := "CTRL+C to EXIT"
	fmt.Fprintf(out, "%*s%s", -targetWidth+len(tooltip), fmt.Sprintf("%*s", (targetWidth+len(text))/2, text), tooltip)
}

func HandleInput(input chan KEY, in *os.File) {
	defer close(input)

	ticker := time.NewTicker(time.Second / 3)
	defer ticker.Stop()

	var lastKey KEY
	var pending bool

	buf := make([]byte, 3)
	for {
		// Read input
		n, err := in.Read(buf)
		if err != nil {
			Log(ERROR, err)
		}

		if n > 0 {
			// Arrow keys
			if n == 3 && buf[0] == 27 && buf[1] == 91 {
				switch buf[2] {
				// Arrow right
				case 67:
					lastKey = RIGHT
					pending = true
				// Arrow left
				case 68:
					lastKey = LEFT
					pending = true
				}
			} else {
				for i := range n {
					char := buf[i]
					// CTRL+C
					if char == 3 {
						lastKey = EXIT
						pending = true
					}
				}
			}
		}

		select {
		case <-ticker.C:
			if pending {
				input <- lastKey
				pending = false
			}
		default:
			continue
		}
	}
}

func WaitAnyKey(in *os.File, out *os.File) {
	inFd := int(in.Fd())

	state, _ := platform.MakeRaw(inFd)
	defer platform.Restore(inFd, state)

	fmt.Fprint(out, "Press any key to exit..."+NEWLINE+RETURN)

	temp := make([]byte, 1)
	n, _ := in.Read(temp)
	if n > 0 {
		return
	}
}

func GetImagesFromDir(dir string) []string {
	images := make([]string, 0)

	entries, err := os.ReadDir(dir)
	if err != nil {
		Log(ERROR, err)
	}

	for _, entry := range entries {

		if entry.IsDir() {
			continue
		}

		switch filepath.Ext(entry.Name()) {
		case ".jpeg", ".jpg", ".png":
			images = append(images, filepath.Join(dir, entry.Name()))
		default:
			continue
		}
	}

	return images
}

func ProcessImage(imagepath string, targetWidth int, scale float64) (*SymbolicImage, error) {
	rawImage, n, err := DecodeImage(imagepath)
	if err != nil {
		return nil, err
	}

	image, err := ScaleImage(rawImage, targetWidth, scale)
	if err != nil {
		return nil, err
	}

	return &SymbolicImage{
		Path:          imagepath,
		Repr:          RenderImage(image),
		Size:          n,
		InitialBounds: rawImage.Bounds(),
		FinalBounds:   image.Bounds(),
	}, nil
}

func ProcessImages(images chan SymbolicImage, imageFilePaths []string, targetWidth int, scale float64) {
	for _, imagefp := range imageFilePaths {
		go func(images chan SymbolicImage) {
			image, err := ProcessImage(imagefp, targetWidth, scale)
			if err == nil {
				images <- *image
			}
		}(images)
	}
}

func InteractiveMode(in *os.File, out *os.File, dir string, targetWidth int, scale float64, extra bool) {
	inFd := int(in.Fd())

	// Change terminal mode to raw
	state, err := platform.MakeRaw(inFd)
	if err != nil {
		Log(ERROR, err)
	}

	defer func() {
		// Change terminal mode back to default
		err = platform.Restore(inFd, state)
		if err != nil {
			Log(ERROR, err)
		}
	}()

	imageFilePaths := GetImagesFromDir(dir)

	images := make([]SymbolicImage, 0)
	imagesCh := make(chan SymbolicImage)

	go ProcessImages(imagesCh, imageFilePaths, targetWidth, scale)

	first := true
	currElem := 1
	maxElems := len(imageFilePaths)
	width := int(float64(targetWidth) * scale)

	inputCh := make(chan KEY, 1)
	go HandleInput(inputCh, in)

	// Hide cursor and show it after exit from interactive mode
	fmt.Fprint(out, CURSOR_HIDE)
	defer fmt.Fprint(out, CURSOR_SHOW)

	for {
		select {
		case key := <-inputCh:
			// Skip input if first image not even processed
			if first {
				continue
			}

			switch key {
			case LEFT:
				currElem = util.Clamp(currElem-1, 1, maxElems)
			case RIGHT:
				// Check bounds
				if currElem+1 <= len(images) {
					currElem = util.Clamp(currElem+1, 1, maxElems)
				}
			case EXIT:
				fmt.Fprint(out, CLEAR)
				return
			}

			img := &images[currElem-1]
			fmt.Print(out, CLEAR)
			fmt.Fprintf(out, "Displaying: %s\n\r", img.Path)
			if extra {
				fmt.Fprint(out, img.FormatExtra())
			}
			fmt.Fprint(out, img.Repr)
			PrintControls(out, currElem, maxElems, width)

		case image := <-imagesCh:
			images = append(images, image)
			// If first image processed - display it
			if first {
				img := &images[currElem-1]
				fmt.Print(out, CLEAR)
				fmt.Fprintf(out, "Displaying: %s\n\r", img.Path)
				if extra {
					fmt.Fprint(out, img.FormatExtra())
				}
				fmt.Fprint(out, img.Repr)
				PrintControls(out, currElem, maxElems, width)
				first = false
			}
		}
	}
}

func main() {
	in := os.Stdin
	out := os.Stdout
	outFd := int(out.Fd())

	flag.Usage = func() {
		fmt.Fprint(out, HELP_TEXT)
		flag.PrintDefaults()
	}

	flag.Parse()

	freezeTerm := *freeze
	extraImageInfo := *extra
	imageRatio := *ratio
	specifiedWidth := *width

	// Enable colored output (platform specific)
	if colorEnabled := platform.EnableColoredOutput(outFd); !colorEnabled {
		Log(ERROR, "colored output not supported")
	}

	// Check if FD is terminal
	isTerminal := platform.IsTerminal(outFd)
	if !isTerminal {
		Log(WARN, "not in terminal")
	}

	// Get terminal width for image scaling
	targetWidth, _, err := platform.GetSize(outFd)
	if err != nil {
		Log(WARN, "could not get terminal bounds")
		if specifiedWidth != 0 {
			Log(WARN, fmt.Sprintf("width specificly set to %d symbols", specifiedWidth))
			targetWidth = specifiedWidth
		} else {
			Log(WARN, "width implicitly set to 100 symbols")
			targetWidth = 100
		}
	}

	imagepath := flag.Arg(0)
	if imagepath == "" {
		Log(ERROR, "image path is empty")
	}

	stat, err := os.Stat(imagepath)
	if err != nil {
		Log(ERROR, err)
	}

	if stat.IsDir() {
		if !isTerminal {
			Log(ERROR, "interactive mode only supported in terminal")
			return
		}
		InteractiveMode(in, out, imagepath, targetWidth, imageRatio, extraImageInfo)
	} else {
		img, err := ProcessImage(imagepath, targetWidth, imageRatio)
		if err != nil {
			Log(ERROR, err)
		}

		if err != nil {
			Log(ERROR, err)
		}

		fmt.Fprint(out, img.Repr)

		if extraImageInfo {
			fmt.Fprint(out, img.FormatExtra())
		}

		if freezeTerm {
			if !isTerminal {
				Log(ERROR, "freeze only supported in terminal")
				return
			}
			WaitAnyKey(in, out)
		}
	}
}
