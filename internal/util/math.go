package util

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint8
}

func Clamp[T Number](value, min, max T) T {
	if value < min {
		return min
	}

	if value > max {
		return max
	}

	return value
}
