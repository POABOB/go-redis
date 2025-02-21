package reply

import (
	"bytes"
	"go-redis/interface/resp"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")
	CRLF               = "\r\n"
)

// ---	A Bulk reply is used to reply a string

type BulkReply struct {
	Arg []byte
}

// ToBytes returns the bytes of bulk with arg and length
func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 {
		return nullBulkReplyBytes
	}
	// E.g. "moody" -> "$5\r\nmoody\r\n"
	return []byte("$" + strconv.Itoa(len(b.Arg)) + CRLF + string(b.Arg) + CRLF)
}

// MakeBulkReply returns an instance of bulk reply
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// --- A Multi-bulk reply is used to return an array of other replies.

type MultiBulkReply struct {
	Args [][]byte
}

// ToBytes returns the bytes of multi-bulk
func (m *MultiBulkReply) ToBytes() []byte {
	// use buffer to make bytes more efficient
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(len(m.Args)) + CRLF)

	for _, arg := range m.Args {
		if arg == nil {
			buf.WriteString(string(nullBulkReplyBytes) + CRLF)
			continue
		}
		buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
	}
	return buf.Bytes()
}

// MakeMultiBulkReply returns an instance of multi-bulk reply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

// --- A Status reply is used to reply a status string

type StatusReply struct {
	Status string
}

// ToBytes returns the bytes of status
func (s *StatusReply) ToBytes() []byte {
	return []byte("+" + s.Status + CRLF)
}

// MakeStatusReply returns an instance of status reply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{Status: status}
}

// --- A Int reply is used to reply a int

type IntReply struct {
	Code int64
}

// ToBytes returns the bytes of int
func (i *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 10) + CRLF)
}

// MakeIntReply returns an instance of status reply
func MakeIntReply(code int64) *IntReply {
	return &IntReply{Code: code}
}

// --- A Standard error reply is used to reply a error

type StandardErrorReply struct {
	Status string
}

// Error returns the error message
func (s *StandardErrorReply) Error() string {
	return s.Status
}

// ToBytes returns the bytes of standard error
func (s *StandardErrorReply) ToBytes() []byte {
	return []byte("-" + s.Status + CRLF)
}

// MakeStandardErrorReply returns an instance of standard error reply
func MakeStandardErrorReply(status string) *StandardErrorReply {
	return &StandardErrorReply{Status: status}
}

// IsErrorReply returns true if the first byte is '-'
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
