package resize

import "math"

type InterpolationFunction func(float64) float64

func LinearInterp(t float64) float64 {
	return t
}

func CubicInterp(t float64) float64 {
	return t * t * (3 - 2*t)
}

func MitchellnetravaliInterp(t float64) float64 {
	t = math.Abs(t)
	if t <= 1 {
		return (7.0*t*t*t - 12.0*t*t + 5.33333333333) * 0.16666666666
	}
	if t <= 2 {
		return (-2.33333333333*t*t*t + 12.0*t*t - 20.0*t + 10.6666666667) * 0.16666666666
	}
	return 0
}

func BiCubicInterp(t float64) float64 {
	if t < 0 {
		t = -t
	}

	if t < 1 {
		return (1.5*t-2.5)*t*t + 1
	}
	if t < 2 {
		return ((-0.5*t+2.5)*t-4)*t + 2
	}
	return 0
}

func LanczosInterp(t float64) float64 {
	if t == 0 {
		return 1
	}
	if t < 0 {
		t = -t
	}
	if t < 3 {
		return math.Sin(math.Pi*t) * math.Sin(math.Pi*t/3) / (math.Pi * math.Pi * t * t / 3)
	}
	return 0
}
