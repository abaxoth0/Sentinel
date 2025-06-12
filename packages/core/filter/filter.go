// Entity filter
package filter

type Condition byte

const (
	Equal Condition = 1 + iota
	Less
	Greater
	LessOrEqual
	GreaterOrEqual
	Like
	IsNull
	IsNotNull
	Contains
	Containd
)

type Entity[P any] struct {
	Property P
	Cond     Condition
	Value    any
}

