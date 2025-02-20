package userdto

func IndexedToUnindexed(indexed *Indexed) *Unindexed {
    return &Unindexed{
        Login: indexed.Login,
        Password: indexed.Password,
        Roles: indexed.Roles,
        DeletedAt: indexed.DeletedAt,
    }
}
