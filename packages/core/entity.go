package core

// Represents property of any entity.
// Must not be used directly, instead create new type definition
// based on this type for each entity and use it.
//
// To avoid possible vulnerabilities like SQL-injections
// all data of types that defined based on this type must be predefined consts.
// Doing so there are no need in properties validations cuz
// all properties are predefined and correct.
type EntityProperty string

