package utils

import "math"

func FloatSave(num float64, n int) float64 {
	if n <= 0 {
		return float64(int64(num))
	}

	return math.Floor(math.Pow(10, float64(n))*num) / math.Pow(10, float64(n))
}
