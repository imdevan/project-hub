package index

import (
	"time"

	"github.com/super-todo/internal/domain"
)

// Indexer is a stub indexer for project scanning.
type Indexer struct {
	store         *Store
	cacheInterval int // seconds
}

// NewIndexer creates an indexer.
func NewIndexer(store *Store, cfg domain.Config) *Indexer {
	return &Indexer{
		store:         store,
		cacheInterval: cfg.CacheInterval,
	}
}

// IndexResult summarises one indexing run.
type IndexResult struct {
	FilesScanned int
	FilesSkipped int
	Duration     time.Duration
	Errors       []error
}

// RefreshAsync performs background refresh if NeedsRefresh returns true.
func (ix *Indexer) RefreshAsync(detached bool) error {
	return nil
}

// Refresh runs refresh tasks.
func (ix *Indexer) Refresh() IndexResult {
	return IndexResult{}
}

// NeedsRefresh reports whether the index needs to be updated.
func (ix *Indexer) NeedsRefresh() bool {
	return false
}
