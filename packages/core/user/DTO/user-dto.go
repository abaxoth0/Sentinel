package userdto

type Indexed struct {
	ID        string   `bson:"_id" json:"_id"`
	Login     string   `bson:"login" json:"login"`
	Password  string   `bson:"password" json:"password"`
	Roles     []string `bson:"roles" json:"roles"`
	DeletedAt int      `bson:"deletedAt,omitmepty" json:"deletedAt"`
}

type Payload struct {
	ID    string   `json:"id"`
	Login string   `json:"login"`
	Roles []string `json:"roles"`
}

type Filter struct {
	TargetUID      string
	RequesterUID   string
	RequesterRoles []string
}

