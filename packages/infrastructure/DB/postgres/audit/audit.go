package audit

type Operation string

const (
	DeleteOperation  Operation = "D"
	UpdatedOperation Operation = "U"
	RestoreOperation Operation = "R"
)

