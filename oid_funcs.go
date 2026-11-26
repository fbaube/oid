package oid

import (
	"encoding/binary"
	"time"
    "uuid"
    "fmt"
    SU "github.com/fbaube/stringutils"
)

// TableName returns the 3-letter DB table code.
// This is non-standard for UUIDs.
func (u *OID) TableName() string {
     return string(u[13:16])
}

// SetTableName sets the 3-letter DB table code.
// This	is non-standard for UUIDs. 
func (u *OID) SetTableName(s string) {
      if len(s) != 3 && len(s) != 0 {
         panic("Nad length of DN table name <" + s + ">")
         }
     // copy(dst []byte, src string) int
     copy(u[13:16], s) 
}

// String assumes a UUID v7 with modifications.
// It renders the OID as "YMDHM-XXX-CCC":
//  - YMDHM is the m5 app's compressed date & time 
//  - XXX is the stdlib's sub-ms ns field (0..4095) 
//  - CCC is the (non-standard!) DB table name 
//
// For the lowercase hex-and-dash representation 
// defined in RFC 9562, use uuid.UUID.String
// .
func (u *OID) String() string {
     // First convert the ms count to a Unix time.
     // Shift the elements into place
     // TODO: Our v7.New does not yet set the ns.
     var unix_ts_ms uint64
     unix_ts_ms = uint64(binary.BigEndian.Uint64(u[:6]) >> 16)
     // timestamp to YMDHM
     // sep. ts bits to sthg else
     // then tbl & rowID
     var unix_time time.Time
     unix_time = time.UnixMilli(int64(unix_ts_ms))
     var ymdhm string
     ymdhm = SU.TimeAsYMDHM(unix_time)
     return ymdhm + fmt.Sprintf("-%1x%2x-%c%c%c", 
            u[6]&0xF, u[7], u[13], u[14], u[15]) 
}

/*
func (u *OID) IsValidKey() error {
    if uuid.UUID(*u) == uuid.Nil() { // NilOID {
        return errors.New("required UUID is Nil (all zeroes)")
    }
    return nil
}
*/
func (u *OID) IsNil() bool {
    return uuid.UUID(*u) == Nil 
}

func (u OIDnilable) OIDisSet() bool {
    return u.IsValid && uuid.UUID(u.OID) != Nil
}
func UUIDasBytes(u any) []byte {
    switch v := u.(type) {
    case uuid.UUID:
        return v[:]
    case OIDnilable:
        if !v.IsValid {
            return nil
        }
        return v.OID[:]
    // TODO: Accept a Nork ? 
    default:
        panic(fmt.Sprintf("UUIDasBytes: bad type: %T", u))
    }
}
func UUIDfromBytes(b []byte) (uuid.UUID, error) { 
     var u uuid.UUID 
     if len(b) != 16 {
        return uuid.Nil(), fmt.Errorf("UUID must be 16 bytes, got %d", len(b))
     }
     // copy(dst []byte, src string) int
     copy(u[:], b)
     return u, nil 
     }
func UUIDisNil(u any) bool {
   switch v := u.(type) {
    case uuid.UUID:
        return v == Nil
    case OIDnilable:
        return !v.IsValid || uuid.UUID(v.OID) == Nil
    default:
        panic(fmt.Sprintf("UUIDisNil: bad type: %T", u))
    }
}
func UUIDasHexString(u uuid.UUID) string { return "UUIDasHexString:TBS" }

func (u *OID) HasValidTableName() bool {
    if uuid.UUID(*u) == Nil { 
        return false 
    }
    got := u.TableName()
    return SU.IsAsciiAlphanum(got) 
}

func (u *OID) NewTabledUUIDv7(tableName string) *OID {
if len(tableName) != 3 { panic("Bad UUISv7 table name: " + tableName) }
     u.SetTableName(tableName)
     return u
}

// Perplexity says: This makes the SQL boundary
// unambiguous: An unset optional UUID becomes SQL
// NULL, while a set UUID becomes a 16-byte BLOB.
func BindNilableOID(u OIDnilable) any {
    if !u.IsValid {
        return nil
    }
    return u.OID[:]
}

