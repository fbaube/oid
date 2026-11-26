package oid

import (
    "uuid"
)

type OID uuid.UUID

type OIDnilable struct {
     IsValid bool
     OID // uuid.UUID
}

