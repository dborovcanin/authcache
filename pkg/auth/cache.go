package auth

import (
	"sync"
	"time"
)

type Cache interface {
	Add(thingID, channelID string)
	Remove(thingID, channelID string)
	Validate(thingID, channelID string) bool
}

type auth struct {
	channelID string
	added     int64
}

type cache struct {
	duration int64
	mu       sync.RWMutex
	data     map[string][]auth
}

func (c *cache) Add(thingID, channelID string) {
	c.mu.Lock()
	d, ok := c.data[thingID]
	if !ok {
		new := auth{
			channelID: channelID,
			added:     time.Now().UnixNano(),
		}
		c.data[thingID] = []auth{new}
		c.mu.Unlock()
		return
	}
	found := -1
	// If already exists, refresh time added and reorder if needed.
	for i, a := range d {
		if a.channelID == channelID {
			found = i
			break
		}
	}
	if found != -1 {
		// Refresh time.
		d[found].added = time.Now().UnixNano()
		// If it's not the first element in the array, make it first.
		if found != 0 {
			switch len(d) {
			case 2:
				d[0], d[1] = d[1], d[0]
			default:
				f := d[found]
				copy(d[1:], append(d[:found], d[found+1:]...))
				d[0] = f
			}
		}
		c.data[thingID] = d
		c.mu.Unlock()
		return
	}
	// Preprend to keep most recently added to the
	// start for quicker getting.
	d = append(d, auth{})
	copy(d[1:], d)
	d[0] = auth{
		channelID: channelID,
		added:     time.Now().UnixNano(),
	}
	c.data[thingID] = d
	c.mu.Unlock()
}

func (c *cache) Remove(thingID, channelID string) {
	c.mu.Lock()
	d, ok := c.data[thingID]
	if !ok {
		c.mu.Unlock()
		return
	}
	found := -1
	for k, v := range d {
		if v.channelID == channelID {
			found = k
			break
		}
	}
	if found == -1 {
		c.mu.Unlock()
		return
	}
	c.data[thingID] = append(d[found:], d[:found+1]...)
	c.mu.Unlock()
}

func (c *cache) Validate(thingID, channelID string) bool {
	c.mu.RLock()
	d, ok := c.data[thingID]
	if !ok {
		c.mu.RUnlock()
		return false
	}

	for i, a := range d {
		if channelID == a.channelID {
			if a.added+c.duration > time.Now().UnixNano() {
				c.mu.RUnlock()
				return true
			}
			// Remove if expired.
			d = d[:i+copy(d[i:], d[i+1:])]
			c.mu.RUnlock()
			return false
		}
	}
	c.mu.RUnlock()
	return false
}
