package email

import (
	Error "sentinel/packages/errors"

)

type Mail interface {
    Send() *Error.Status
}

