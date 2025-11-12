//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package tag

import "github.com/iomz/go-llrp"

// ManagementAction represents the type of operation to perform on tags.
type ManagementAction int

const (
	// RetrieveTags retrieves all currently stored tags.
	RetrieveTags ManagementAction = iota
	// AddTags adds new tags to the collection (duplicates are ignored).
	AddTags
	// DeleteTags removes specified tags from the collection.
	DeleteTags
)

// Manager represents a tag management command sent through the management channel.
// It specifies the action to perform and the tags to operate on.
type Manager struct {
	Action ManagementAction // The operation to perform
	Tags   llrp.Tags        // Tags to add, delete, or empty for retrieve
}
