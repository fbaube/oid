package oid

import (
    "uuid"
)

type OID uuid.UUID

// Nil is used as a flag value, for 
// example for an empty foreign key
// UUID (i.e. a link to another object). 
var Nil = uuid.Nil() 

