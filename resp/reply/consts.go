package reply

// --- A Pong reply is used to return from ping.

type PongReply struct {
}

var (
	pongBytes    = []byte("+PONG\r\n")
	thePongReply = &PongReply{}
)

// ToBytes returns the bytes of pong
func (p *PongReply) ToBytes() []byte {
	return pongBytes
}

// MakePongReply returns an instance of pong reply
func MakePongReply() *PongReply {
	return thePongReply
}

// --- A Ok reply is used to reply ok.

type OkReply struct {
}

var (
	okBytes    = []byte("+OK\r\n")
	theOkReply = &OkReply{}
)

// ToBytes returns the bytes of ok
func (o *OkReply) ToBytes() []byte {
	return okBytes
}

// MakeOkReply returns an instance of ok reply
func MakeOkReply() *OkReply {
	return theOkReply
}

// --- A Null bulk reply is used to return an empty string.

type NullBulkReply struct {
}

var (
	nullBulkBytes    = []byte("$-1\r\n")
	theNullBulkReply = &NullBulkReply{}
)

// ToBytes returns the bytes of empty bulk
func (p *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

// MakeNullBulkReply returns an instance of empty bulk reply
func MakeNullBulkReply() *NullBulkReply {
	return theNullBulkReply
}

// --- A Multi bulk reply is used to return an array of other replies.

type MultiBulkReply struct {
}

var (
	multiBulkBytes    = []byte("*0\r\n")
	theMultiBulkReply = &MultiBulkReply{}
)

// ToBytes returns the bytes of empty multi-bulk
func (m *MultiBulkReply) ToBytes() []byte {
	return multiBulkBytes
}

// MakeMultiBulkReply returns an instance of empty multi-bulk reply
func MakeMultiBulkReply() *MultiBulkReply {
	return theMultiBulkReply
}

// --- A No reply is used to return when a command is not found.

type NoReply struct {
}

var (
	noReplyBytes = []byte("")
	theNoReply   = &NoReply{}
)

// ToBytes returns the bytes of empty
func (n *NoReply) ToBytes() []byte {
	return noReplyBytes
}

// MakeNoReply returns an instance of no reply
func MakeNoReply() *NoReply {
	return theNoReply
}
