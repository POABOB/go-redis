package reply

// --- A UnknownErrorReply is used to reply unknown error

var (
	unknownErrorBytes    = []byte("-ERR unknown\r\n")
	theUnknownErrorReply = &UnknownErrorReply{}
)

type UnknownErrorReply struct {
}

// Error returns the error message
func (u *UnknownErrorReply) Error() string {
	return "ERR unknown"
}

// ToBytes returns the bytes of unknown
func (u *UnknownErrorReply) ToBytes() []byte {
	return unknownErrorBytes
}

// MakeUnknownErrorReply returns a instance of UnknownErrorReply
func MakeUnknownErrorReply() *UnknownErrorReply {
	return theUnknownErrorReply
}

// --- A SyntaxErrorReply is used to reply syntax error

var (
	syntaxErrorBytes    = []byte("-ERR syntax error\r\n")
	theSyntaxErrorReply = &SyntaxErrorReply{}
)

type SyntaxErrorReply struct {
}

// Error returns the error message
func (s *SyntaxErrorReply) Error() string {
	return "ERR syntax error"
}

// ToBytes returns the bytes of syntax error
func (s *SyntaxErrorReply) ToBytes() []byte {
	return syntaxErrorBytes
}

// MakeSyntaxErrorReply returns a instance of SyntaxErrorReply
func MakeSyntaxErrorReply() *SyntaxErrorReply {
	return theSyntaxErrorReply
}

// --- A WrongTypeErrorReply is used to reply wrong type

var (
	wrongTypeErrorBytes    = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
	theWrongTypeErrorReply = &WrongTypeErrorReply{}
)

type WrongTypeErrorReply struct {
}

// ToBytes returns the bytes of wrong type
func (w *WrongTypeErrorReply) ToBytes() []byte {
	return wrongTypeErrorBytes
}

// Error returns the error message
func (w *WrongTypeErrorReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}

// MakeWrongTypeErrorReply returns a instance of WrongTypeErrorReply
func MakeWrongTypeErrorReply() *WrongTypeErrorReply {
	return theWrongTypeErrorReply
}

// --- A ProtocolErrorReply is used to reply protocol error

type ProtocolErrorReply struct {
	Message string
}

// ToBytes returns the bytes of protocol error
func (p *ProtocolErrorReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + p.Message + "'\r\n")
}

// Error returns the error message
func (p *ProtocolErrorReply) Error() string {
	return "ERR Protocol error: '" + p.Message
}

// MakeProtocolErrorReply returns a new instance of ProtocolErrorReply
func MakeProtocolErrorReply(message string) *ProtocolErrorReply {
	return &ProtocolErrorReply{Message: message}
}

// --- A ArgsNumErrorReply is used to reply wrong number of arguments

type ArgsNumErrorReply struct {
	Command string
}

// Error returns the error message
func (a *ArgsNumErrorReply) Error() string {
	return "-ERR wrong number of arguments for '" + a.Command + "' command\r\n"
}

// ToBytes returns the bytes of wrong number of arguments error
func (a *ArgsNumErrorReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + a.Command + "' command\r\n")
}

// MakeArgsNumErrorReply returns a new instance of ArgsNumErrorReply
func MakeArgsNumErrorReply(command string) *ArgsNumErrorReply {
	return &ArgsNumErrorReply{Command: command}
}
