// This package defines the Store interfaces for reading and writing bundles
// (or any artifact data) to some underlying system, and Store implementations.
package store

import (
	"io"
	"time"

	"github.com/twitter/scoot/common/stats"
)

var DefaultTTL time.Duration = time.Hour * 24 * 180 //180 days. If zero, no ttl will be applied by default.
const DefaultTTLKey string = "x-scoot-expires"      //the primary use for this is communicating ttl(RFC1123) over http.

// Stores should generally support TTL, at this time only httpStore implements it.
type TTLValue struct {
	TTL    time.Time
	TTLKey string
}

type TTLConfig struct {
	TTL    time.Duration
	TTLKey string
}

// Gets a TTLValue based on Now given a TTLConfig, or nil
func GetTTLValue(c *TTLConfig) *TTLValue {
	if c != nil {
		return &TTLValue{TTL: time.Now().Add(c.TTL), TTLKey: c.TTLKey}
	}
	return nil
}

// Read-only operations on store, limited for now to a couple essential functions.
type StoreRead interface {
	// Check if the bundle exists. Not guaranteed to be any cheaper than actually reading the bundle.
	Exists(name string) (bool, error)

	// Open the bundle for streaming read. It is the caller's responsibility to call Close().
	OpenForRead(name string) (io.ReadCloser, error)

	// Get the base location, like a directory or base URI that the Store writes to
	Root() string
}

// Write operations on store, limited to a one-shot writing operation since bundles are immutable.
// If ttl config is nil then the store will use its defaults.
type StoreWrite interface {
	// Does a streaming write of the given bundle. There is no concept of partial writes (partial=failed).
	Write(name string, data io.Reader, ttl *TTLValue) error
}

// Combines read and write operations on store. This is what most of the code will use.
type Store interface {
	StoreRead
	StoreWrite
}

// Encapsulating struct for instances of Stores and accompanying configurations
type StoreConfig struct {
	Store  Store
	TTLCfg *TTLConfig
	Stat   stats.StatsReceiver
}
