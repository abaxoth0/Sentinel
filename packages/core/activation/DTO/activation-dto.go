package activationdto

import "time"

type Basic struct {
    Id int
    UserLogin string
    Token string
    ExpiresAt time.Time
    CreatedAt time.Time
}

