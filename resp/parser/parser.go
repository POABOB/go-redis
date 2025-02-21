package parser

import (
	"bufio"
	"errors"
	"go-redis/interface/resp"
	"io"
)

type Payload struct {
	Data  resp.Reply // resp.Reply is both from client and server
	Error error
}

type readState struct {
	readingMultiLine  bool // reading multi line or single line
	expectedArgsCount int  // expected how many args we need
	messageType       byte
	args              [][]byte // args of command from client
	bulkLength        int64    // count the bulk length
}

// isFinished returns true if the args count is equal to the expected args count
func (r *readState) isFinished() bool {
	return r.expectedArgsCount > 0 && r.expectedArgsCount == len(r.args)
}

// ParseStream returns a received-only channel of Payload to make it parallel parsing
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// parse0 only send payload to channel
func parse0(reader io.Reader, ch chan<- *Payload) {

}

/**
 * readLine reads message from client and below are the two rules to read:
 * 1. read the line by CRLF
 * 2. parse the bytes number with ${number} or *${number}
 *
 * if the readLine occurs IO error, the bool value is true
 * otherwise, the bool value is false when other error occurs
 */
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var message []byte
	var err error
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

// isMessageComplete returns true if the message is end with CRLF
func isMessageComplete(message []byte) bool {
	return len(message) > 0 && message[len(message)-2] == '\r' && message[len(message)-1] == '\n'
}
