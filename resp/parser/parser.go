package parser

import (
	"bufio"
	"errors"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
	"io"
	"strconv"
	"strings"
)

type Payload struct {
	Data  resp.Reply // resp.Reply is both from client and server
	Error error
}

type readState struct {
	readingMultiLine  bool     // reading multi line or single line
	expectedArgsCount int      // expected how many args we need
	messageType       byte     // E.g. '*', '-', '$', '+', ':'
	args              [][]byte // args of command from client
	bulkLength        int64    // count the bulk length
}

// isFinished returns true if the args count is equal to the expected args count.
func (r *readState) isFinished() bool {
	return r.expectedArgsCount > 0 && r.expectedArgsCount == len(r.args)
}

// ParseStream returns a received-only channel of Payload to make it parallel parsing.
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// parse0 only send payload to channel.
func parse0(reader io.Reader, ch chan<- *Payload) {

}

/*
*
readLine reads message from client and below are the two rules to read:
1. read the line by CRLF
2. parse the bytes number with ${number} or *${number}

if the readLine occurs IO error, the bool value is true.
otherwise, the bool value is false when other error occurs.
*/
func readLine(bufReader *bufio.Reader, state *readState) (message []byte, isIOError bool, err error) {
	if state.bulkLength == 0 { // use method 1.
		message, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if isMessageComplete(message) {
			return nil, false, errors.New("protocol error: " + string(message))
		}
	} else { // use method 2.
		message = make([]byte, state.bulkLength+2) // +2 for CRLF
		_, err = io.ReadFull(bufReader, message)
		if err != nil {
			return nil, true, err
		}
		if isMessageComplete(message) {
			return nil, false, errors.New("protocol error: " + string(message))
		}
		state.bulkLength = 0 // reset
	}
	return message, false, nil
}

// isMessageComplete returns true if the message is end with CRLF.
func isMessageComplete(message []byte) bool {
	return len(message) > 0 && message[len(message)-2] == '\r' && message[len(message)-1] == '\n'
}

// parseMultiBulkHeader parses the multi bulk header and init the readState.
// E.g. "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n".
func parseMultiBulkHeader(message []byte, state *readState) (err error) {
	state.bulkLength, err = strconv.ParseInt(string(message[1:len(message)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(message))
	}

	if state.bulkLength == 0 {
		state.expectedArgsCount = 0
		return
	} else if state.bulkLength < 0 {
		err = errors.New("protocol error: " + string(message))
		return
	}
	state.messageType = message[0]
	state.readingMultiLine = true
	state.expectedArgsCount = int(state.bulkLength)
	state.args = make([][]byte, 0, state.bulkLength)
	return
}

// parseBulkHeader parses the bulk header and init the readState.
// E.g. "$4\r\nPING\r\n".
func parseBulkHeader(message []byte, state *readState) (err error) {
	state.bulkLength, err = strconv.ParseInt(string(message[1:len(message)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(message))
	}

	if state.bulkLength == -1 { // null bulk
		return
	} else if state.bulkLength > 0 {
		state.messageType = message[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
	} else {
		err = errors.New("protocol error: " + string(message))
	}
	return
}

// parseSingleLineReply parses the single line and make reply.
// E.g. "+OK\r\n" or "-err\r\n" or ":5\r\n".
func parseSingleLineReply(message []byte) (result resp.Reply, err error) {
	str := strings.TrimSuffix(string(message), "\r\n")
	switch message[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeStandardErrorReply(str[1:])
	case ':':
		var value int64
		value, err = strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			err = errors.New("protocol error: " + string(message))
			return
		}
		result = reply.MakeIntReply(value)
	}
	return
}

// readBody reads the body of the message.
// E.g. "$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n".
func readBody(message []byte, state readState) (err error) {
	line := message[:len(message)-2] // exclude CRLF
	if line[0] == '$' {
		state.bulkLength, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			err = errors.New("protocol error: " + string(message))
			return
		}
		if state.bulkLength <= 0 { // $0\r\n
			state.args = append(state.args, []byte{})
			state.bulkLength = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return
}
