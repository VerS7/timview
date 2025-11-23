package resize

import (
	"image"
	"image/color"
	"math"

	"timview/internal/util"
)

func Resize(width, height uint, img image.Image, interp InterpolationFunction, samples int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if width == 0 || height == 0 {
		if width == 0 && height == 0 {
			return img
		}
		if width == 0 {
			ratio := float64(srcWidth) / float64(srcHeight)
			width = uint(float64(height) * ratio)
		} else {
			ratio := float64(srcHeight) / float64(srcWidth)
			height = uint(float64(width) * ratio)
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	scaleX := float64(srcWidth) / float64(width)
	scaleY := float64(srcHeight) / float64(height)

	if samples > 1 {
		intermediateWidth := srcWidth / samples
		intermediateHeight := srcHeight / samples

		if intermediateWidth > 0 && intermediateHeight > 0 {
			intermediate := smooth(uint(intermediateWidth), uint(intermediateHeight), img, interp, 2)
			return smooth(width, height, intermediate, interp, 2)
		}
	}

	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			// Multi-sampling
			var r, g, b, a float64

			for sy := 0; sy < samples; sy++ {
				for sx := 0; sx < samples; sx++ {
					offsetX := (float64(sx) / float64(samples)) - 0.5
					offsetY := (float64(sy) / float64(samples)) - 0.5

					srcX := (float64(x) + offsetX) * scaleX
					srcY := (float64(y) + offsetY) * scaleY

					c := bicubicInterpolate(img, srcX, srcY, interp, 2)
					cr, cg, cb, ca := c.RGBA()

					r += float64(cr >> 8)
					g += float64(cg >> 8)
					b += float64(cb >> 8)
					a += float64(ca >> 8)
				}
			}

			samplesSq := float64(samples * samples)
			dst.Set(x, y, color.RGBA{
				R: uint8(util.Clamp(r/samplesSq, 0, 255)),
				G: uint8(util.Clamp(g/samplesSq, 0, 255)),
				B: uint8(util.Clamp(b/samplesSq, 0, 255)),
				A: uint8(util.Clamp(a/samplesSq, 0, 255)),
			})
		}
	}

	return dst
}

func smooth(width, height uint, img image.Image, interp InterpolationFunction, radius float64) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if width == 0 || height == 0 {
		if width == 0 && height == 0 {
			return img
		}
		if width == 0 {
			ratio := float64(srcWidth) / float64(srcHeight)
			width = uint(float64(height) * ratio)
		} else {
			ratio := float64(srcHeight) / float64(srcWidth)
			height = uint(float64(width) * ratio)
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	scaleX := float64(srcWidth) / float64(width)
	scaleY := float64(srcHeight) / float64(height)

	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			srcX := float64(x) * scaleX
			srcY := float64(y) * scaleY

			color := bicubicInterpolate(img, srcX, srcY, interp, radius)
			dst.Set(x, y, color)
		}
	}

	return dst
}

func bicubicInterpolate(img image.Image, x, y float64, interp InterpolationFunction, radius float64) color.Color {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))

	var r, g, b, a float64
	var totalWeight float64

	bounds := img.Bounds()

	for i := -int(radius); i <= int(radius)+1; i++ {
		for j := -int(radius); j <= int(radius)+1; j++ {
			srcX := x0 + i
			srcY := y0 + j

			if srcX < bounds.Min.X || srcX >= bounds.Max.X ||
				srcY < bounds.Min.Y || srcY >= bounds.Max.Y {
				continue
			}

			dx := (x - float64(srcX)) / radius
			dy := (y - float64(srcY)) / radius

			weightX := interp(math.Abs(dx))
			weightY := interp(math.Abs(dy))
			weight := weightX * weightY

			// Получаем цвет
			c := img.At(srcX, srcY)
			cr, cg, cb, ca := c.RGBA()

			r += float64(cr>>8) * weight
			g += float64(cg>>8) * weight
			b += float64(cb>>8) * weight
			a += float64(ca>>8) * weight
			totalWeight += weight
		}
	}

	if totalWeight > 0 {
		r /= totalWeight
		g /= totalWeight
		b /= totalWeight
		a /= totalWeight
	}

	return color.RGBA{
		R: uint8(util.Clamp(r, 0, 255)),
		G: uint8(util.Clamp(g, 0, 255)),
		B: uint8(util.Clamp(b, 0, 255)),
		A: uint8(util.Clamp(a, 0, 255)),
	}
}
