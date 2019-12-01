package sessions

import (
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/go-redis/redis"
)

// RedisStore represents a session.Store backed by redis.
type RedisStore struct {
	Client          *redis.Client
	SessionDuration time.Duration
}

// NewRedisStore constructs a new RedisStore
func NewRedisStore(client *redis.Client, sessionDuration time.Duration) *RedisStore {
	return &RedisStore{client, sessionDuration}
}

// Store implementation

// Save saves the provided `sessionState` and associated SessionID to the store.
// The `sessionState` parameter is typically a pointer to a struct containing
// all the data you want to be associated with the given SessionID.
func (rs *RedisStore) Save(sid SessionID, sessionState interface{}) error {
	//TODO: marshal the `sessionState` to JSON and save it in the redis database,
	//using `sid.getRedisKey()` for the key.
	//return any errors that occur along the way.

	// turn sessionState to JSON
	buffer, err := json.Marshal(sessionState)
	if err != nil {
		return err
	}

	err = rs.Client.Set(sid.getRedisKey(), buffer, 0).Err()
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

// Get populates `sessionState` with the data previously saved
// for the given SessionID
func (rs *RedisStore) Get(sid SessionID, sessionState interface{}) error {
	val, err := rs.Client.Get(sid.getRedisKey()).Result()
	if err != nil {
		return ErrStateNotFound
	}
	// Turn the value into the sessionState decoded JSON parameter
	buffer := []byte(val)
	if err := json.Unmarshal(buffer, sessionState); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
		return err
	}

	err = rs.Client.Set(sid.getRedisKey(), val, 0).Err()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Delete deletes all state data associated with the SessionID from the store.
func (rs *RedisStore) Delete(sid SessionID) error {
	err := rs.Client.Del(sid.getRedisKey()).Err()
	if err != nil {
		return ErrStateNotFound
	}

	return nil
}

// getRedisKey returns the redis key to use for the SessionID
func (sid SessionID) getRedisKey() string {
	return "sid:" + sid.String()
}
