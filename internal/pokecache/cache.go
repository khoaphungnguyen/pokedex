package pokecache

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type Cache struct {
	cache          map[string]cacheEntry
	mu               sync.RWMutex 
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
}

type cacheEntry struct {
	createdAt   time.Time
	val         []byte
	expireAfter time.Duration
}

// NewCache creates and returns a new Cache instance with a cleanup goroutine.
// cleanupInterval specifies how frequently expired entries should be removed.
// defaultExpiration specifies the default expiration duration for each entry unless overridden.
func NewCache(cleanupInterval, defaultExpiration time.Duration) *Cache {
	c := &Cache{
		cache:             make(map[string]cacheEntry),
		cleanupInterval:   cleanupInterval,
		defaultExpiration: defaultExpiration,
	}
	// Start the cleanup goroutine to remove expired entries periodically.
	go c.cleanupExpiredEntries()
	return c
}

func (c *Cache) Set(key string, value []byte, expireAfter ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Duration
	if len(expireAfter) > 0 {
		expiration = expireAfter[0]
	} else {
		expiration = c.defaultExpiration
	}

	c.cache[key] = cacheEntry{
		createdAt:   time.Now(),
		val:         value,
		expireAfter: expiration,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock() 
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists || (entry.expireAfter != 0 && time.Since(entry.createdAt) > entry.expireAfter) {
		return nil, false
	}

	return entry.val, true
}

// cleanupExpiredEntries periodically checks and removes expired entries from the cache.
func (c *Cache) cleanupExpiredEntries() {
	for {
		time.Sleep(c.cleanupInterval)
		c.mu.Lock()
		for key, entry := range c.cache {
			if entry.expireAfter != 0 && time.Since(entry.createdAt) > entry.expireAfter {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Iterate(f func(key string, value []byte) bool) {
    c.mu.RLock() // Use read lock since we're only reading the cache
    defer c.mu.RUnlock()

    for key, entry := range c.cache {
        // Call the passed function `f` for each cache entry.
        // If `f` returns false, stop the iteration early.
        if !f(key, entry.val) {
            break
        }
    }
}

// SavePokedex saves the caught Pokémon to a file.
func (c *Cache) SavePokedex(filepath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	caughtPokemons := make(map[string][]byte)
	for key, entry := range c.cache {
		if strings.HasPrefix(key, "caught:") {
			caughtPokemons[key] = entry.val
		}
	}

	data, err := json.Marshal(caughtPokemons)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath, data, 0644)
}

// LoadPokedex loads the caught Pokémon from a file.
func (c *Cache) LoadPokedex(filepath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File not found is not an error
		}
		return err
	}

	var caughtPokemons map[string][]byte
	if err := json.Unmarshal(data, &caughtPokemons); err != nil {
		return err
	}

	for key, val := range caughtPokemons {
		c.cache[key] = cacheEntry{
			createdAt:   time.Now(),
			val:         val,
			expireAfter: 100 * 365 * 24 * time.Hour, // Setting a long expiration time
		}
	}

	return nil
}