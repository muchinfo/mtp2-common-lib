package utils

type PointNumber interface {
	int | int32 | bool | float64 | uint | uint32 | string
}

// SetPointValue 给Struct的指针类型成员赋值
func SetPointValue[T PointNumber](value T) (field *T) {
	field = new(T)
	*field = value

	return
}
