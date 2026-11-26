package oid

import (
    "uuid"
//  SU "github.com/fbaube/stringutils"
)

type OID uuid.UUID

//  NilOID is an empty UUID, the zero value, all zeros.
var NilOID OID // = (uuid.UUID).(OID)

//  MaxOID has all 128 bits set to 1.
var MaxOID = OID(uuid.UUID{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	})

type OIDnilable struct {
     IsValid bool
     OID // uuid.UUID
}

