package vec

import (
	"math"
)

type (
	Vector3 struct {
		X float32
		Y float32
		Z float32
	}
)

var (
	Vector3Zero = Vector3{}
	Vector3One  = Vector3{X: 1, Y: 1, Z: 1}
)

func CreateVector3(x, y, z float32) Vector3 {
	return Vector3{
		X: x,
		Y: y,
		Z: z,
	}
}
func (vec3 Vector3) GetX() float32 {
	return vec3.X
}
func (vec3 Vector3) GetY() float32 {
	return vec3.Y
}
func (vec3 Vector3) GetZ() float32 {
	return vec3.Z
}

func (vec3 Vector3) Normalize() Vector3 {
	num := vec3.Magnitude()
	if num > 1e-05 {
		return vec3.DivNumber(num)
	}
	return Vector3Zero
}

// 加
func (vec3 Vector3) Add(v Vector3) Vector3 {
	return Vector3{
		X: vec3.X + v.X,
		Y: vec3.Y + v.Y,
		Z: vec3.Z + v.Z,
	}
}

// 除
func (vec3 Vector3) AddNumber(v float32) Vector3 {
	return Vector3{
		X: vec3.X + v,
		Y: vec3.Y + v,
		Z: vec3.Z + v,
	}
}

// 减
func (vec3 Vector3) Sub(v Vector3) Vector3 {
	return Vector3{
		X: vec3.X - v.X,
		Y: vec3.Y - v.Y,
		Z: vec3.Z - v.Z,
	}
}

// 除
func (vec3 Vector3) SubNumber(v float32) Vector3 {
	return Vector3{
		X: vec3.X - v,
		Y: vec3.Y - v,
		Z: vec3.Z - v,
	}
}

// 乘
func (vec3 Vector3) Mul(v Vector3) Vector3 {
	return Vector3{
		X: vec3.X * v.X,
		Y: vec3.Y * v.Y,
		Z: vec3.Z * v.Z,
	}
}

// 除
func (vec3 Vector3) MulNumber(v float32) Vector3 {
	return Vector3{
		X: vec3.X * v,
		Y: vec3.Y * v,
		Z: vec3.Z * v,
	}
}

// 除
func (vec3 Vector3) Div(v Vector3) Vector3 {
	return Vector3{
		X: vec3.X / v.X,
		Y: vec3.Y / v.Y,
		Z: vec3.Z / v.Z,
	}
}

// 除
func (vec3 Vector3) DivNumber(v float32) Vector3 {
	return Vector3{
		X: vec3.X / v,
		Y: vec3.Y / v,
		Z: vec3.Z / v,
	}
}

func (vec3 Vector3) ConvertVec2() Vector2 {
	return Vector2{
		X: vec3.X,
		Y: vec3.Y,
	}
}

func (vec3 Vector3) SqrMagnitude() float32 {
	return vec3.X*vec3.X + vec3.Y*vec3.Y + vec3.Z*vec3.Z
}
func (vec3 Vector3) Magnitude() float32 {
	return float32(math.Sqrt(float64(vec3.SqrMagnitude())))
}

func (vec3 Vector3) Distance(target Vector3) float32 {
	return vec3.Sub(target).Magnitude()
}
