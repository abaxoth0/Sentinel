package role

type Permission string

func (p Permission) String() string {
	return string(p)
}

const CreatePermission Permission = "C"
const SelfCreatePermission Permission = "SC"

const ReadPermission Permission = "R"
const SelfReadPermission Permission = "SR"

const UpdatePermission Permission = "U"
const SelfUpdatePermission Permission = "SU"

const DeletePermission Permission = "D"
const SelfDeletePermission Permission = "SD"

const ModeratorPermission Permission = "M"
const AdminPermission Permission = "A"

var Permissions []Permission = []Permission{
	CreatePermission,
	SelfCreatePermission,
	ReadPermission,
	SelfReadPermission,
	UpdatePermission,
	SelfUpdatePermission,
	DeletePermission,
	SelfDeletePermission,
	ModeratorPermission,
	AdminPermission,
}
