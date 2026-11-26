package oid

import (
    "uuid"
)

type OID uuid.UUID

// Nil is used as a flag value, for example
// for an empty link to another object. 
var Nil = uuid.Nil() 

type OIDnilable struct {
     IsValid bool
     OID // uuid.UUID
}
