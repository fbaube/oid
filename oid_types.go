package oid

import (
    "uuid"
)

type OID uuid.UUID

// Nil is used as a flag value, for 
// example for an empty foreign key
// UUID (i.e. a link to another object). 
var Nil = uuid.Nil() 

/*
Perplexity suggests some tests to verify that...
- Generate two OIDs in a tight loop; their timestamps 
  must be nondecreasing, and the OIDs must differ.
- Bytes  8–12 are not all zero across several generated OIDs.
- Bytes 13–15 equal the requested table code.
- String() does not panic and that its first 48 timestamp 
  bits decode to approximately time.Now() 
- NewFromBytes rejects 15- and 17-byte slices.
- OID{} is Nil, IsNil() is true, and HasValidTableName() is false.
*/

