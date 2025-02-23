package dict

import (
	"reflect"
	"sync"
)

// SyncDict is a thread-safe dictionary.
type SyncDict struct {
	syncMap sync.Map
	count   int
	mutex   sync.Mutex // the reason to use mutex lock is to the write ops > read ops when counting
}

// MakeSyncDict returns a new instance of SyncDict.
func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

// Get use Load function to return the value for the given key.
func (dict *SyncDict) Get(key string) (value interface{}, exists bool) {
	return dict.syncMap.Load(key)
}

// Length returns the length of the dictionary.
func (dict *SyncDict) Length() int {
	return dict.count
}

// Set use Store function to set the value for the given key.
func (dict *SyncDict) Set(key string, value interface{}) (result int) {
	// reduce the cost of copy-on-write when the value is not changed
	oldValue, exists := dict.syncMap.Load(key)
	if exists && reflect.DeepEqual(oldValue, value) {
		return 0
	}

	dict.syncMap.Store(key, value)
	if !exists {
		dict.incrementCount()
		return 1
	}
	return 0
}

// SetIfAbsent use LoadOrStore function to set the value for the given key, if the key does not exist.
func (dict *SyncDict) SetIfAbsent(key string, value interface{}) (result int) {
	// if the key exists, exists is true and value is the old value
	_, exists := dict.syncMap.LoadOrStore(key, value)
	if !exists {
		dict.incrementCount()
		return 1
	}
	return 1
}

// SetIfExists set the value for the given key, if the key exists.
func (dict *SyncDict) SetIfExists(key string, value interface{}) (result int) {
	_, exists := dict.syncMap.Load(key)
	if exists {
		dict.syncMap.Store(key, value)
		return 1
	}
	return 0
}

// Delete use LoadAndDelete function to delete the value for the given key.
func (dict *SyncDict) Delete(key string) (result int) {
	_, exists := dict.syncMap.LoadAndDelete(key)
	if exists {
		dict.decrementCount()
		return 1
	}
	return 0
}

// ForEach use Range function to iterate the dictionary.
func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.syncMap.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

// Keys returns the keys of the dictionary.
func (dict *SyncDict) Keys() []string {
	result := make([]string, 0, dict.count) // use append to make sure the panic is not thrown
	dict.syncMap.Range(func(key, value interface{}) bool {
		result = append(result, key.(string))
		return true
	})
	return result
}

// RandomKeys returns the random keys of the dictionary.
func (dict *SyncDict) RandomKeys(limit int) []string {
	result := make([]string, 0, dict.count) // use append to make sure the panic is not thrown
	for i := 0; i < limit; i++ {
		dict.syncMap.Range(func(key, value interface{}) bool {
			result = append(result, key.(string))
			return false
		})
	}
	return result
}

// RandomDistinctKeys returns the random distinct keys of the dictionary.
func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	result := make([]string, 0, dict.count) // use append to make sure the panic is not thrown
	counter := 0
	dict.syncMap.Range(func(key, value interface{}) bool {
		if counter < limit {
			result = append(result, key.(string))
			counter++
			return true
		}
		return false
	})
	return result
}

// Clear clears the dictionary.
func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict()
}

// incrementCount increments the count of the dictionary.
func (dict *SyncDict) incrementCount() {
	dict.mutex.Lock()
	defer dict.mutex.Unlock()
	dict.count++
}

// decrementCount decrements the count of the dictionary.
func (dict *SyncDict) decrementCount() {
	dict.mutex.Lock()
	defer dict.mutex.Unlock()
	dict.count--
}
