# go-redis

`RESP (Redis Serialization Protocol)` is a protocol for communication between clients and servers, with five types of message formats:
- `Simple Strings`: Start with "+" and end with "\r\n".
- `Errors`: Start with "-" and end with "\r\n".
- `Integers`: Start with ":" and end with "\r\n".
- `Bulk Strings`: Start with "$", followed by the actual byte length, and enclosed with "\r\n".
- `Multi Bulk Strings`: Start with "*", followed by the actual number of bulk strings, and enclosed with "\r\n".