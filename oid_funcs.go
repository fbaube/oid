package oid

import (
	"encoding/binary"
	"time"
    "uuid"
    "fmt"
    SU "github.com/fbaube/stringutils"
)

func (u *OID) UUID() uuid.UUID {
      return uuid.UUID(*u)
}

// TableName returns the 3-letter DB table name. 
// This is non-standard for UUIDs.
func (u *OID) TableName() string {
     return string(u[13:16])
}

// SetTableName sets the 3-letter DB table name. 
// This	is non-standard for UUIDs. 
func (u *OID) SetTableName(s string) {
      if s == "" {
         clear(u[13:16])
         return
         }
      if len(s) != 3 {
         panic("Bad length for OID's DB table name <" + s + ">")
         }
     if !SU.IsAsciiAlphanum(s) {
        panic("Bad chars for OID's DB table name <" + s + ">")
     }
     // copy(dst []byte, src string) int
     copy(u[13:16], s) 
}

// String assumes a UUID v7 with modifications.
// It renders the OID as "YMDHM-XXX-CCC":
//  - YMDHM is the m5 app's compressed date & time
//    (e.g. 6bpi0 = 2026, 11=Nov, i = 18th, 0=00:00) 
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
 var unix_ts_ms uint64 // Perplexity suggests a fix 
  // unix_ts_ms = uint64(binary.BigEndian.Uint64(u[:6]) >> 16) 
     unix_ts_ms = binary.BigEndian.Uint64(u[0:8]) >> 16
     // timestamp to YMDHM
     // sep. ts bits to sthg else
     // then tbl & rowID
     var unix_time time.Time
     unix_time = time.UnixMilli(int64(unix_ts_ms))
     var ymdhm string
     ymdhm = SU.TimeAsYMDHM(unix_time)
     return ymdhm + fmt.Sprintf("-%1x%02x-%c%c%c", 
            u[6]&0xF, u[7], u[13], u[14], u[15]) 
}

func (u *OID) IsNil() bool {
    return uuid.UUID(*u) == Nil 
}
func (u *OID) IsSet() bool {
    return uuid.UUID(*u) != Nil 
}

func (u *OID) Bytes() []byte {
      // return u[:] // len 16 // Perplexity says this just aliases
      b := make([]byte, 16)
      copy(b, u[:])
      return b
}

func (u *OID) HasValidTableName() bool {
    if uuid.UUID(*u) == Nil { 
        return false 
    }
    got := u.TableName()
    return SU.IsAsciiAlphanum(got) 
}

