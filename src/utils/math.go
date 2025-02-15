package utils

import "math"

func FloatSave(num float64, n int) (res float64) {
	defer func() {
		// 有除法，防止零除
		if r := recover(); r != nil {
			res = 0.0
		}
	}()

	if num == 0 {
		return num
	}

	if n <= 0 {
		return float64(int64(num))
	}

	return math.Floor(math.Pow(10, float64(n))*num) / math.Pow(10, float64(n))
}
