package dict

import (
	"hash/fnv"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

// numShards is the number of shards to use in the sharded dictionary.
const numShards = 16

// ShardedDict is a thread-safe dictionary using sharding for performance.
type ShardedDict struct {
	shards []*syncMapShard
	count  int
	mutex  sync.Mutex // Protects count, which is shared across all shards
	random *rand.Rand
}

// syncMapShard holds a single shard's sync.Map and a mutex for synchronization.
type syncMapShard struct {
	syncMap sync.Map
	mutex   sync.Mutex
}

// MakeShardedDict returns a new instance of ShardedDict.
func MakeShardedDict() *ShardedDict {
	shards := make([]*syncMapShard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = &syncMapShard{}
	}
	source := rand.NewSource(time.Now().UnixNano())
	return &ShardedDict{shards: shards, random: rand.New(source)}
}

// shardForKey returns the shard index for a given key.
func (dict *ShardedDict) shardForKey(key string) *syncMapShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	shardIndex := h.Sum32() % uint32(numShards)
	return dict.shards[shardIndex]
}

// Get use Load function to return the value for the given key.
func (dict *ShardedDict) Get(key string) (value interface{}, exists bool) {
	shard := dict.shardForKey(key)
	// prevent race condition for concurrent access the shard
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	return shard.syncMap.Load(key)
}

// Length returns the total length of all shards combined.
func (dict *ShardedDict) Length() int {
	dict.mutex.Lock()
	defer dict.mutex.Unlock()
	return dict.count
}

// Set use Store function to set the value for the given key.
func (dict *ShardedDict) Set(key string, value interface{}) (result int) {
	shard := dict.shardForKey(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	oldValue, exists := shard.syncMap.Load(key)
	if exists && reflect.DeepEqual(oldValue, value) {
		return 0
	}

	shard.syncMap.Store(key, value)
	if !exists {
		dict.incrementCount()
		return 1
	}
	return 0
}

// SetIfAbsent use LoadOrStore function to set the value for the given key, if the key does not exist.
func (dict *ShardedDict) SetIfAbsent(key string, value interface{}) (result int) {
	shard := dict.shardForKey(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	// LoadOrStore function to set the value if the key does not exist
	_, exists := shard.syncMap.LoadOrStore(key, value)
	if !exists {
		dict.incrementCount()
		return 1
	}
	return 0
}

// SetIfExists set the value for the given key, if the key exists.
func (dict *ShardedDict) SetIfExists(key string, value interface{}) (result int) {
	shard := dict.shardForKey(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.syncMap.Load(key)
	if exists {
		shard.syncMap.Store(key, value)
		return 1
	}
	return 0
}

// Delete use LoadAndDelete function to delete the value for the given key.
func (dict *ShardedDict) Delete(key string) (result int) {
	shard := dict.shardForKey(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, exists := shard.syncMap.LoadAndDelete(key)
	if exists {
		dict.decrementCount()
		return 1
	}
	return 0
}

// ForEach use Range function to iterate the dictionary.
func (dict *ShardedDict) ForEach(consumer Consumer) {
	for _, shard := range dict.shards {
		shard.mutex.Lock()
		shard.syncMap.Range(func(key, value interface{}) bool {
			consumer(key.(string), value)
			return true
		})
		shard.mutex.Unlock()
	}
}

// Keys returns the keys of the dictionary.
func (dict *ShardedDict) Keys() []string {
	result := make([]string, 0, dict.count)
	for _, shard := range dict.shards {
		shard.mutex.Lock()
		shard.syncMap.Range(func(key, value interface{}) bool {
			result = append(result, key.(string))
			return true
		})
		shard.mutex.Unlock()
	}
	return result
}

// RandomKeys returns the random keys of the dictionary.
func (dict *ShardedDict) RandomKeys(limit int) []string {
	return dict.getRandomKeys(limit, false)
}

// RandomDistinctKeys returns random distinct keys of the dictionary.
func (dict *ShardedDict) RandomDistinctKeys(limit int) []string {
	return dict.getRandomKeys(limit, true)
}

// getRandomKeys processes the random keys.
func (dict *ShardedDict) getRandomKeys(limit int, isDistinct bool) []string {
	// Ensure we only seed once at the start of the program
	if dict.random == nil {
		dict.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	result := make([]string, 0, limit)
	seen := make(map[string]struct{})
	for limit > len(result) {
		// Randomly select a shard
		shard := dict.shards[rand.Intn(numShards)]
		shard.mutex.Lock()

		// Randomly pick a number of keys to fetch from this shard (between 1 and 10)
		numKeysToFetch := rand.Intn(10) + 1 // Random number between 1 and 10

		// Iterate through the shard's keys
		shard.syncMap.Range(func(key, value interface{}) bool {
			if len(result) >= limit {
				return false
			}
			// Only add the key if it's not already in the result
			if isDistinct {
				if _, found := seen[key.(string)]; !found {
					seen[key.(string)] = struct{}{}
					result = append(result, key.(string))
					numKeysToFetch--
				}
			} else {
				result = append(result, key.(string))
				numKeysToFetch--
			}
			return limit > len(result) && numKeysToFetch > 0
		})
		shard.mutex.Unlock()
	}
	return result
}

// Clear clears all shards in the dictionary.
func (dict *ShardedDict) Clear() {
	dict.mutex.Lock()
	defer dict.mutex.Unlock()

	for _, shard := range dict.shards {
		shard.mutex.Lock()
		shard.syncMap = sync.Map{} // Reset the shard's map
		shard.mutex.Unlock()
	}

	// Reset the count as well
	dict.count = 0
}

// incrementCount increments the count of the dictionary.
func (dict *ShardedDict) incrementCount() {
	dict.mutex.Lock()
	defer dict.mutex.Unlock()
	dict.count++
}

// decrementCount decrements the count of the dictionary.
func (dict *ShardedDict) decrementCount() {
	dict.mutex.Lock()
	defer dict.mutex.Unlock()
	dict.count--
}
