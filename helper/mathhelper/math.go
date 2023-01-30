package mathhelper

import "math"

//Distance 两点之间距离
func Distance(x1, y1, x2, y2 float32) float32 {
	return float32(math.Hypot(float64(x2-x1), float64(y2-y1)))
	// return float32(math.Pow(float64(x1-x2)*float64(x1-x2)+float64(y1-y2)*float64(y1-y2), 0.5))
}
