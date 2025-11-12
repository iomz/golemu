//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package tag

import (
	"sync"
	"sync/atomic"

	"github.com/iomz/go-llrp"
)

// ManagerService handles tag management operations
type ManagerService struct {
	tags           llrp.Tags
	tagManagerChan chan Manager
	tagUpdatedChan chan llrp.Tags
	isConnAlive    *atomic.Bool
	mu             sync.Mutex
}

// NewManagerService creates a new tag manager service
func NewManagerService(tagManagerChan chan Manager, tagUpdatedChan chan llrp.Tags, isConnAlive *atomic.Bool) *ManagerService {
	return &ManagerService{
		tags:           llrp.Tags{},
		tagManagerChan: tagManagerChan,
		tagUpdatedChan: tagUpdatedChan,
		isConnAlive:    isConnAlive,
	}
}

// Process handles tag management commands
func (s *ManagerService) Process(cmd Manager) {
	var tagsToNotify llrp.Tags
	var shouldNotify bool

	s.mu.Lock()
	res := []*llrp.Tag{}
	switch cmd.Action {
	case AddTags:
		for _, t := range cmd.Tags {
			if i := s.tags.GetIndexOf(t); i < 0 {
				s.tags = append(s.tags, t)
				res = append(res, t)
			}
		}
		if len(res) > 0 && s.isConnAlive.Load() {
			// Make a copy of tags before releasing the lock
			tagsToNotify = make(llrp.Tags, len(s.tags))
			copy(tagsToNotify, s.tags)
			shouldNotify = true
		}
	case DeleteTags:
		for _, t := range cmd.Tags {
			if i := s.tags.GetIndexOf(t); i >= 0 {
				s.tags = append(s.tags[:i], s.tags[i+1:]...)
				res = append(res, t)
			}
		}
		if len(res) > 0 && s.isConnAlive.Load() {
			// Make a copy of tags before releasing the lock
			tagsToNotify = make(llrp.Tags, len(s.tags))
			copy(tagsToNotify, s.tags)
			shouldNotify = true
		}
	case RetrieveTags:
		res = s.tags
	}
	cmd.Tags = res
	s.mu.Unlock()

	// Send to channels without holding the lock to avoid deadlock
	if shouldNotify {
		s.tagUpdatedChan <- tagsToNotify
	}
	s.tagManagerChan <- cmd
}

// GetTags returns the current tags
func (s *ManagerService) GetTags() llrp.Tags {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tags
}

// SetTags sets the tags
func (s *ManagerService) SetTags(tags llrp.Tags) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tags = tags
}
