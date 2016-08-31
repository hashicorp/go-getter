package getter

import (
	"sync"
)

// Group represents a group of related downloads.
//
// By using a Group with a Client, go-getter can optimize downloads:
// if a URL was downloaded previously (to a completely different path),
// go-getter will instead copy this download rather than re-download it.
//
// Additionally, when using a group with parallelized downloads, the clients
// will automatically wait for related downlodas to complete. For example, if
// you start two downloads of the same Git URL in parallel with a client
// configured with a Group, then one of the downloads will simply wait for
// the other to complete and copy that data.
//
// Preconditions: A group assumes that the downloads happen separate from
// use because the group may copy files from the download directories and
// they're expected to not modify from underneath it. Therefore, it is up
// to the consumer of this functionality to ensure that all grouped downloads
// happen prior to any use.
type Group struct {
	// cache is the cached data for a group. The key is a URL String() value
	// and the value is the
	entries     map[string]*groupEntry
	entriesLock sync.RWMutex
}

// Begin is called to signal that a download is beginning for the given URL.
//
// begin returns the location where this URL has already been downloaded
// (if it exists). If it hasn't been downloaded, the return value is empty.
//
// This function itself doesn't download anything. It is called internally
// by the Client to note a started operation.
func (g *Group) Begin(k string) string {
	entry := g.entry(k)
	entry.Lock.Lock()
	defer entry.Lock.Unlock()

	// If this is already started, we wait for the operation to finish
	for entry.Began {
		entry.Cond.Wait()
	}

	// We may or may not have a location that we've already downloaded to.
	// Regardless, we mark that this entry has began and return. The contract
	// of this API states that `end` must be called even for `begin` calls
	// that return a cached location.
	entry.Began = true
	return entry.Location
}

// End is called and paired with a begin call to signal that the downloading
// of a certain key (likely a URL) is complete.
//
// End must be called for EVERY Begin call, even if the begin call returns
// a location.
//
// End must not be called if Begin was not called. Calling End without
// a matching Begin will panic the application.
func (g *Group) End(k string, result string) {
	entry := g.entry(k)
	entry.Lock.Lock()
	defer entry.Lock.Unlock()

	// Against the API contract
	if !entry.Began {
		panic("end() called without begin()")
	}

	// If we have a location to set, set it. We allow "" to not set the
	// location to make the API easier to use in certain cases (such as
	// errors).
	if result != "" {
		entry.Location = result
	}

	entry.Began = false
	entry.Cond.Broadcast()
}

// entry retrieves the entry for the given key.
func (g *Group) entry(k string) *groupEntry {
	g.entriesLock.Lock()
	defer g.entriesLock.Unlock()

	// Initialize if we must
	if g.entries == nil {
		g.entries = make(map[string]*groupEntry)
	}

	// Get the entry from the map. If it doesn't exist, create it.
	e, ok := g.entries[k]
	if !ok {
		e = new(groupEntry)
		e.Cond = sync.NewCond(&e.Lock)

		g.entries[k] = e
	}

	return e
}

// groupEntry represents a single entry in the list of downloads within
// a Group. These should never be constructed mnaully, they are constructed
// by a Group.
type groupEntry struct {
	Location string     // Location of the entry data
	Began    bool       // If true, someone else is trying to download right now
	Cond     *sync.Cond // Condition variable to wait for data
	Lock     sync.Mutex // Mutex to modify this entry
}
