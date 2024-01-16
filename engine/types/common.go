package types

type Number interface {
	int | uint | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64
}

type Vector2 interface {
	GetX() float32
	GetY() float32
}
type Vector3 interface {
	GetX() float32
	GetY() float32
	GetZ() float32
}
