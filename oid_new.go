package oid

import(
    "fmt"
    "time"
    "sync"
    "math/rand"
    "encoding/binary"
    SU "github.com/fbaube/stringutils"
)

// q.v. https://www.rfc-editor.org/rfc/rfc9562.html

// These need to be package scope ! 
   var (
    oidMutex         sync.Mutex
    oidLastTimestamp uint64
//  oidLastEpochSecs uint64
)

// OID is a UUID that borrows from the Go stdlib UUID
// version 7 a bit, and veers from the RFC a bit. 
// It is not field-for-field compatible with RFC 9562
// standard UUID v7, but it is fully interoperable. 
//  - Input argument "tableName", the DB table name, 
//    must be either three letters or the empty string 
//    (""). NOTE: To favor UC or LC (or neither) is TBD. 
//  - Version 7 UUIDs have a timestamp in the first 
//    48 bits, and at least 62 bits of random data
//  - func NewV7 returns UUIDs that sort in increasing 
//    order, except if the system clock moves backwards
//  - Go stdlib UUID v7 overwrites field "rand_a" with 
//    a nanoseconds field, as permitted by the v7 spec
//  - Contrary to spec, OID overwrites the last 24 bits 
//    of field "rand_b" with a three-letter DB table code, 
//    so that each OID  is a self-contained DB reference - 
//    except that the three-letter code must (tipicly) be 
//    translated to a full DB table name.
//  - These two changes reduce the random in the UUID from a
//    possible max  of 74 bits down to an effective 38 bits.
//
// Note that OIDs are comparable using the "==" operator,
// because a UUID is an array (not a slice). "==" WORKS!
// .
func NewV7(tableName string) OID {

// OID: 
// unix_ts_sub_ms (formerly: rand_a) 
// is 12 bits of sub-ms timestamp, 
// as optionally permitted by the RFC.
// rand_b is reduced to 38 bits of random. 
// unix_ts_ms is 48 bits i.e. bytes [0..5]
// ver is the hi-niblle of byte[6.]
// var is the hi-two-bits of byte[8]. 
//
//   bits 0                   1                   2                   3
//        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//  bytes +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  0..3  |                           unix_ts_ms                          |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  4..7  |          unix_ts_ms           |  ver  |    unix_ts_sub_ms     |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  8..11 |var|                        rand_b                             |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// 12..15 |    rand_b     |      'T'      |      'B'      |      'L'      |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// We store a 12 bit sub-millisecond timestamp fraction in 
// the rand_a section, as optionally permitted by the RFC.
// 60 bits in our timestamp:
// 48 bits of millisecond-resolution, followed by
// 12 bits of 1/4096-millisecond resolution.
//
// RFC 9562 section 5.7 UUID v7:
//
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           unix_ts_ms                          |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |          unix_ts_ms           |  ver  |       rand_a          |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |var|                        rand_b                             |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                            rand_b                             |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

     if tableName != "" && len(tableName) != 3 {
        panic("Bad length of OID table name <" + tableName + ">")
        }
     if !SU.IsAsciiAlphanum(tableName) {
        panic("Bad chars in OID table name <" + tableName + ">")
        }
   // =========
   //  PROCEED 
   // =========
	oidMutex.Lock()
    var now time.Time 
    now = time.Now()
    // func (t Time) Unix() int64 ::
    // The number of seconds elapsed
    // since (start of) 1970.01.01 UTC .
    // Currently 1.8 billion (needs 31 bits) 
	epochSecs := uint64(now.Unix())
    // func (t Time) Nanosecond() int :: (ns mod s) :: 
    // The nanosecond offset within the second specified
    // by t, in the range [0, (10**9)-1] (needs 30 bits) 
	nanos := uint64(now.Nanosecond())
    // (ms mod s) :: The millisecond offset within the 
    // second specified by t, in the range [0, 999] (10 bits)
	msecs := nanos / 1_000_000
    // (ns mod ms): The nanosecond offset within the MILLIsecond 
    // specified by t, in the range [0, (10**6)-1] (needs 30 bits)
	fracns   :=  nanos - (1_000_000 * msecs)
    fracns4k := (fracns * 4096) / 1_000_000 
    // So now we have:
    // epochSecs (31 bits) , msecs (10 bits) , fracns , fracns4k (12 bits)
    // This result needs (31+10+12) bits = 53 bits 
	timestamp := ((1000*epochSecs + msecs) << 12) + fracns4k 

    // Here are two IF's that have the same code
    // but differing inline documentation !
    // 
    // System time has gone backwards ?
    if timestamp < oidLastTimestamp {
       timestamp = oidLastTimestamp + 1
		// Set our copy of the clock back, altho 
        // this ignores (and in theory could cause 
        // conflicts with) previously-generated UUIDs.
        // TODO: Put a srsly grumpy message on os.Stderr ?
    //        
	// Traffic jam ?
	} else if timestamp == oidLastTimestamp {
		// Traffic jam!
        // To preserve the property that we generate OIDs
        // in order, use a timestamp 1/4096 millisecond 
        // later than the most recently generated UUID.
        // NOTE: This places a soft limit on sustained 
        // thruput of 4000+ UUIDs per second. 
		timestamp = oidLastTimestamp + 1
	}
	// oidLastEpochSecs = epochSecs
	oidLastTimestamp = timestamp
	oidMutex.Unlock()

    // Perplexity says: 
    millis := timestamp >> 12
    frac := timestamp & 0x0fff
    var u OID
    binary.BigEndian.PutUint64(
        u[0:8], (millis<<16)|(uint64(7)<<12)|uint64(frac))
    rand.Read(u[8:16])
    /*
    The resulting byte layout is:
    bytes 0..5   48-bit millisecond timestamp   
    byte  6 hi   version 7
    byte  6 lo   sub-ms fraction, bits 11..8
    byte  7      sub-ms fraction, bits 7..0
    byte  8 hi   variant 10
    bytes 8..12  random
    bytes 13..15 table code
    
    rand.Read(u[8:16]) is also correct now.
    It fills all five random bytes, and SetTableName
    then replaces only the final three.
    */
    // Perplexity end 

    // Perplexity is replacing this block:
/*
	// Insert a gap for the 4 bits of the ver field into the timestamp.
	hibbytes := ((timestamp << 4) &        // room for one hex digit 
                  0xffff_ffff_ffff_0000) | // 12 hex digits = 48 bits 
                 (timestamp & 0x0ffff)     //  4 hex digits = 16 bits
    // hibytes is 10 bytes and now (in hex) TTTT TTTT TTTT vvvv TTTT

    // ==============
    //  NOW BUILD IT
    // ==============
    var u OID 
    // var  BigEndian bigEndian
    // type bigEndian struct{}
    // func PutUint64 stores v into b[0:8].
    // func (bigEndian) PutUint64(b []byte, v uint64) {
    // BigEndian is the big-endian implementation
    // of ByteOrder and AppendByteOrder.
    // So this loads 8 bytes i.e. [0..7], ?? omitting two bytes +? 
	binary.BigEndian.PutUint64(u[0:8], hibbytes)
    // rand.Read generates len(p) random bytes from 
    // the default [Source] and writes them into p. 
    // It always returns len(p) and a nil error.
    // It is safe for concurrent use.
    // func Read(p []byte) (n int, err error) {
    //      return globalRand().Read(p) }
	// rand.Read(u[8:])
    // Now instead of 8 bytes of random, we load only one.
	rand.Read(u[8:8]) // FIXME FIXME // ?? rand.Read(u[8:16]) ?? 
*/
	u.setVersion(7)
	u.setVariant(0b10)
    // =======================
    //  INSERT the TABLE NAME
    // =======================
     u.SetTableName(tableName)
     return u
}

// setVersion zeroes the high hex digit and then
// sets it to the version (left-shifted ofc). 
func (u *OID) setVersion(version byte) {
	u[6] = (u[6] & 0b0000_1111) | (version << 4)
}

// setVariant zeroes the high two bits and then
// sets themn to the variant (left-shifted ofc).
func (u *OID) setVariant(variant byte) {
	u[8] = (u[8] & 0b0011_1111) | (variant << 6)
}

func NewFromBytes(b []byte) (OID, error) {
     var u OID
     if len(b) != 16 {
        return OID(Nil), fmt.Errorf("UUID must be 16 bytes, got %d", len(b))
     }
     // copy(dst []byte, src byte[]) int
     copy(u[:], b)
     return u, nil
     }

