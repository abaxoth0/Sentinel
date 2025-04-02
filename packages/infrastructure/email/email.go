package email

import Error "sentinel/packages/common/errors"

type Email interface {
    Send() *Error.Status
}

