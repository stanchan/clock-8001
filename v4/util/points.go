package util

import (
	"math"
)

// Point is a simple point in a 2D plane
type Point struct {
	X int32
	Y int32
}

// Points generates a circle center at (center, center) with radius r and count points
// points are returned clockwise array
func Points(center int32, r float64, count int) []Point {
	points := make([]Point, count) // Temporary untill we rearrange them
	for i := 0; i < count; i++ {
		points[i] = Point{
			X: center - int32(r*math.Sin(-float64(i)*2*math.Pi/float64(count))),
			Y: center - int32(r*math.Cos(-float64(i)*2*math.Pi/float64(count))),
		}
	}
	return points
}
