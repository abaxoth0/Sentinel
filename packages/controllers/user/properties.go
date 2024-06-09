package user

type property string

const emailProperty property = "email"
const passwordProperty property = "password"
const roleProperty property = "role"
const deletedAtProperty property = "deletedAt"

var userProperties [4]property = [4]property{emailProperty, passwordProperty, roleProperty, deletedAtProperty}

func (property property) Verify() bool {
	for _, validProperty := range userProperties {
		if property == validProperty {
			return true
		}
	}

	return false
}
