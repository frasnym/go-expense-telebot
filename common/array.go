package common

func InsertAndShift[T any](slice []T, element T) []T {
	return append([]T{element}, slice...)
}
