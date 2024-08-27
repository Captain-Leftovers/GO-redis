package resp

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StoreEntry struct {
	value      string
	expiration time.Time
}

var store = make(map[string]StoreEntry)
var mu sync.RWMutex

var commands = map[string]func(...BulkString) ([]byte, error){
	"ECHO": handleEcho,
	"PING": handlePing,
	"SET":  handleSet,
	"GET":  handleGet,
}

func StartCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		var keysToDelete []string

		mu.RLock()
		now := time.Now()
		for key, entry := range store {
			if now.After(entry.expiration) {
				keysToDelete = append(keysToDelete, key)
			}
		}
		mu.RUnlock()

		if len(keysToDelete) > 0 {
			mu.Lock()
			for _, key := range keysToDelete {
				if entry, exists := store[key]; exists && now.After(entry.expiration) {
					delete(store, key)
				}
			}
			mu.Unlock()
		}
	}
}

// handleSet sets the value of the key in a map with an optional expiration time in milliseconds and returns OK
// if no time is provided the key will expire in 24 hours
func handleSet(params ...BulkString) ([]byte, error) {
	len := len(params)
	if len < 3 {

		return []byte(""), fmt.Errorf("invalid argument for SET")
	}

	key := params[1]
	value := params[2]

	if len > 3 {
		flagOne := *params[3].Value
		if strings.ToUpper(flagOne) == "PX" && len > 4 {
			// check if the value is an integer because PX expects time in milliseconds
			pxValue, err := strconv.Atoi(*params[4].Value)
			if err != nil {
				return []byte(""), fmt.Errorf("invalid argument for SET")
			}

			mu.Lock()
			store[*key.Value] = StoreEntry{value: *value.Value, expiration: time.Now().Add(time.Duration(pxValue) * time.Millisecond)}

			mu.Unlock()
			return []byte("+OK\r\n"), nil
		}

	}

	mu.Lock()
	store[*key.Value] = StoreEntry{value: *value.Value, expiration: time.Now().Add(24 * time.Hour)}
	mu.Unlock()
	return []byte("+OK\r\n"), nil
}

// handleGet returns the value of the key stored in a map if it exists or nil BulkString if it doesn't
func handleGet(args ...BulkString) ([]byte, error) {
	key := *args[1].Value

	// Step 1: Acquire read lock to check the existence and expiration of the key
	mu.RLock()
	storeEntry, exists := store[key]
	mu.RUnlock() // Release read lock as we may need to acquire a write lock

	// Step 2: Handle key existence and expiration
	if exists {
		// If the entry has expired, we need to delete it
		if time.Now().After(storeEntry.expiration) {
			// Step 3: Acquire write lock to delete the expired key
			mu.Lock()
			// Double-check the condition to ensure it hasn't been modified
			if storeEntry, exists := store[key]; exists && time.Now().After(storeEntry.expiration) {
				delete(store, key)
				mu.Unlock() // Release write lock after deletion
				return []byte("$-1\r\n"), nil
			}
			mu.Unlock() // Release write lock if no deletion occurred
		} else {
			// Key is valid, return its value
			return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(storeEntry.value), storeEntry.value)), nil
		}
	}

	// Key does not exist or has been deleted, return null bulk string
	return []byte("$-1\r\n"), nil
}

// handlePing returns PONG
func handlePing(args ...BulkString) ([]byte, error) {
	return []byte("+PONG\r\n"), nil
}

// handleEcho returns the second argument as a response
func handleEcho(args ...BulkString) ([]byte, error) {
	if len(args) < 2 {
		return []byte(""), fmt.Errorf("missing argument for ECHO")
	}

	return args[1].serialize()
}

// parses the RESP data checks for commands and returns serialized response string and error if any
func ExecuteRespData(data []byte) ([]byte, error) {
	respData, _, err := ParseByteDataToResp(data)
	if err != nil {
		return []byte(""), err
	}

	val, ok := respData.(Array)
	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	// TODO : see if we can use more than 2 elements
	if len(*val.Elements) < 1 {
		return []byte(""), fmt.Errorf("invalid command")
	}

	cmd, ok := (*val.Elements)[0].(BulkString)

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	cmdStr := strings.ToUpper(*cmd.Value)

	commandFunc, ok := commands[cmdStr]

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	if len(*val.Elements) < 2 {
		return commandFunc()
	}

	args := make([]BulkString, len(*val.Elements))

	for i, elem := range *val.Elements {
		arg, ok := elem.(BulkString)

		if !ok {
			return []byte(""), fmt.Errorf("invalid command")
		}

		args[i] = arg
	}

	// arg, ok := (*val.Elements)[1].(BulkString)

	if !ok {
		return []byte(""), fmt.Errorf("invalid command")
	}

	return commandFunc(args...)

}
