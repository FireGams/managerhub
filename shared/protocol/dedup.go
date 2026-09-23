package protocol

import "sync"

// Deduper drops duplicate message IDs within a sliding window.
// It is safe for concurrent use.
type Deduper struct {
	mu   sync.Mutex
	seen map[string]struct{}
	order []string
	max  int
}

// NewDeduper creates a Deduper keeping up to max IDs.
func NewDeduper(max int) *Deduper {
	if max <= 0 {
		max = 4096
	}
	return &Deduper{seen: make(map[string]struct{}, max), max: max}
}

// Seen reports whether id was already observed and records it otherwise.
// Empty IDs are never considered duplicates.
func (d *Deduper) Seen(id string) bool {
	if id == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.seen[id]; ok {
		return true
	}
	d.seen[id] = struct{}{}
	d.order = append(d.order, id)
	if len(d.order) > d.max {
		old := d.order[0]
		d.order = d.order[1:]
		delete(d.seen, old)
	}
	return false
}
