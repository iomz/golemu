//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package tag

import "github.com/iomz/go-llrp"

// ManagementAction is a type for TagManager
type ManagementAction int

const (
	// RetrieveTags is a const for retrieving tags
	RetrieveTags ManagementAction = iota
	// AddTags is a const for adding tags
	AddTags
	// DeleteTags is a const for deleting tags
	DeleteTags
)

// Manager is a struct for tag management channel
type Manager struct {
	Action ManagementAction
	Tags   llrp.Tags
}
