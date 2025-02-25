package dict

// Consumer is a callback function, if return true, it will continue to iterate
type Consumer func(key string, value interface{}) bool

type Dict interface {
	Get(key string) (value interface{}, exists bool)
	GetAndDelete(key string) (value interface{}, exists bool)
	Length() int
	Set(key string, value interface{}) (result int)
	SetIfAbsent(key string, value interface{}) (result int)
	SetIfExists(key string, value interface{}) (result int)
	Delete(key string) (result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
}
