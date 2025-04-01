package email

import Error "sentinel/packages/common/errors"

type Mail interface {
    Send() *Error.Status
}

