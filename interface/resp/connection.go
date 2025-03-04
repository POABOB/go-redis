package resp

type Connection interface {
	Write([]byte) error
	GetDBIndex() int
	SelectDB(int)

	SetPassword(string)
	GetPassword() string
}
