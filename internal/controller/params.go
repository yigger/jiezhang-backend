package controller

type invalidParamError struct {
	field string
}

func (e invalidParamError) Error() string {
	return "invalid parameter: " + e.field
}

func ErrInvalidParam(field string) error {
	return invalidParamError{field: field}
}
