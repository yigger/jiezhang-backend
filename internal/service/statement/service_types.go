package statement

import (
	"fmt"
)

type ValidateError struct {
	Message string
}

func (e ValidateError) Error() string {
	return fmt.Sprintf("validate error: %s", e.Message)
}

type GetCategoriesInput struct {
	AccountBookID int64
	Type          string
}
