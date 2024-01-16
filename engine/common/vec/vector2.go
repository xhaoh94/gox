package vec

import (
	"math"
)

type (
	Vector2 struct {
		X float32
		Y float32
	}
)

var (
	Vector2Zero = Vector2{X: 0, Y: 0}
	Vector2One  = Vector2{X: 1, Y: 1}
)

func CreateVector2(x, y float32) Vector2 {
	return Vector2{
		X: x,
		Y: y,
	}
}

func (vec2 Vector2) GetX() float32 {
	return vec2.X
}
func (vec2 Vector2) GetY() float32 {
	return vec2.Y
}

func (vec2 Vector2) Normalize() Vector2 {
	num := vec2.Magnitude()
	if num > 1e-05 {
		return vec2.DivNumber(num)
	}
	return Vector2Zero
}

// 加
func (vec2 Vector2) Add(v Vector3) Vector2 {
	return Vector2{
		X: vec2.X + v.X,
		Y: vec2.Y + v.Y,
	}
}
func (vec2 Vector2) AddNumber(v float32) Vector2 {
	return Vector2{
		X: vec2.X + v,
		Y: vec2.Y + v,
	}
}

// 减
func (vec2 Vector2) Sub(v Vector2) Vector2 {
	return Vector2{
		X: vec2.X - v.X,
		Y: vec2.Y - v.Y,
	}
}
func (vec2 Vector2) SubNumber(v float32) Vector2 {
	return Vector2{
		X: vec2.X - v,
		Y: vec2.Y - v,
	}
}

// 乘
func (vec2 Vector2) Mul(v Vector2) Vector2 {
	return Vector2{
		X: vec2.X * v.X,
		Y: vec2.Y * v.Y,
	}
}
func (vec2 Vector2) MulNumber(v float32) Vector2 {
	return Vector2{
		X: vec2.X * v,
		Y: vec2.Y * v,
	}
}

// 除
func (vec2 Vector2) Div(v Vector2) Vector2 {
	return Vector2{
		X: vec2.X / v.X,
		Y: vec2.Y / v.Y,
	}
}
func (vec2 Vector2) DivNumber(v float32) Vector2 {
	return Vector2{
		X: vec2.X / v,
		Y: vec2.Y / v,
	}
}

func (vec2 Vector2) ConvertVec3() Vector3 {
	return Vector3{
		X: vec2.X,
		Y: vec2.Y,
		Z: 0,
	}
}
func (vec2 Vector2) Distance(target Vector2) float32 {
	return vec2.Sub(target).Magnitude()
}
func (vec2 Vector2) SqrMagnitude() float32 {
	return vec2.X*vec2.X + vec2.Y*vec2.Y
}
func (vec2 Vector2) Magnitude() float32 {
	return float32(math.Sqrt(float64(vec2.SqrMagnitude())))
}
