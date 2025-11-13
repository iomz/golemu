//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package tag

import (
	"sync/atomic"
	"testing"

	"github.com/iomz/go-llrp"
)

func TestNewManagerService(t *testing.T) {
	tagManagerChan := make(chan Manager, 1)
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	if service == nil {
		t.Fatal("NewManagerService returned nil")
	}
	if service.tagManagerChan != tagManagerChan {
		t.Error("tagManagerChan not set correctly")
	}
	if service.tagUpdatedChan != tagUpdatedChan {
		t.Error("tagUpdatedChan not set correctly")
	}
	if service.isConnAlive != isConnAlive {
		t.Error("isConnAlive not set correctly")
	}
	if len(service.tags) != 0 {
		t.Error("tags should be empty initially")
	}
}

func TestManagerService_GetTags(t *testing.T) {
	tagManagerChan := make(chan Manager, 1)
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tags := service.GetTags()
	if tags == nil {
		t.Error("GetTags should not return nil")
	}
	if len(tags) != 0 {
		t.Error("GetTags should return empty tags initially")
	}
}

func TestManagerService_SetTags(t *testing.T) {
	tagManagerChan := make(chan Manager, 1)
	tagUpdatedChan := make(chan llrp.Tags, 1)
	isConnAlive := &atomic.Bool{}

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	tag2, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101011"})
	if err != nil {
		t.Fatalf("failed to create tag2: %v", err)
	}

	tags := llrp.Tags{tag1, tag2}
	service.SetTags(tags)

	retrieved := service.GetTags()
	if len(retrieved) != 2 {
		t.Errorf("expected 2 tags, got %d", len(retrieved))
	}
}

func TestManagerService_Process_AddTags(t *testing.T) {
	tagManagerChan := make(chan Manager, 10)
	tagUpdatedChan := make(chan llrp.Tags, 10)
	isConnAlive := &atomic.Bool{}
	isConnAlive.Store(true)

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	cmd := Manager{
		Action: AddTags,
		Tags:   llrp.Tags{tag1},
	}

	service.Process(cmd)

	// Check response
	select {
	case result := <-tagManagerChan:
		if len(result.Tags) != 1 {
			t.Errorf("expected 1 tag in result, got %d", len(result.Tags))
		}
		if result.Tags[0] != tag1 {
			t.Error("returned tag does not match added tag")
		}
	default:
		t.Error("no response received on tagManagerChan")
	}

	// Check tag updated channel
	select {
	case updatedTags := <-tagUpdatedChan:
		if len(updatedTags) != 1 {
			t.Errorf("expected 1 tag in updated tags, got %d", len(updatedTags))
		}
	default:
		t.Error("no update sent on tagUpdatedChan")
	}

	// Verify tag was added
	tags := service.GetTags()
	if len(tags) != 1 {
		t.Errorf("expected 1 tag in service, got %d", len(tags))
	}
}

func TestManagerService_Process_AddTags_Duplicate(t *testing.T) {
	tagManagerChan := make(chan Manager, 10)
	tagUpdatedChan := make(chan llrp.Tags, 10)
	isConnAlive := &atomic.Bool{}
	isConnAlive.Store(true)

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Add tag first time
	cmd1 := Manager{Action: AddTags, Tags: llrp.Tags{tag1}}
	service.Process(cmd1)
	<-tagManagerChan // consume response
	<-tagUpdatedChan // consume update

	// Try to add same tag again
	cmd2 := Manager{Action: AddTags, Tags: llrp.Tags{tag1}}
	service.Process(cmd2)

	select {
	case result := <-tagManagerChan:
		if len(result.Tags) != 0 {
			t.Errorf("expected 0 tags in result (duplicate), got %d", len(result.Tags))
		}
	default:
		t.Error("no response received on tagManagerChan")
	}

	// Should not send update for duplicate
	select {
	case <-tagUpdatedChan:
		t.Error("should not send update for duplicate tag")
	default:
		// Expected - no update
	}

	// Verify only one tag exists
	tags := service.GetTags()
	if len(tags) != 1 {
		t.Errorf("expected 1 tag in service, got %d", len(tags))
	}
}

func TestManagerService_Process_DeleteTags(t *testing.T) {
	tagManagerChan := make(chan Manager, 10)
	tagUpdatedChan := make(chan llrp.Tags, 10)
	isConnAlive := &atomic.Bool{}
	isConnAlive.Store(true)

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Add tag first
	addCmd := Manager{Action: AddTags, Tags: llrp.Tags{tag1}}
	service.Process(addCmd)
	<-tagManagerChan
	<-tagUpdatedChan

	// Delete tag
	deleteCmd := Manager{Action: DeleteTags, Tags: llrp.Tags{tag1}}
	service.Process(deleteCmd)

	select {
	case result := <-tagManagerChan:
		if len(result.Tags) != 1 {
			t.Errorf("expected 1 tag in result, got %d", len(result.Tags))
		}
	default:
		t.Error("no response received on tagManagerChan")
	}

	// Check tag updated channel
	select {
	case updatedTags := <-tagUpdatedChan:
		if len(updatedTags) != 0 {
			t.Errorf("expected 0 tags after deletion, got %d", len(updatedTags))
		}
	default:
		t.Error("no update sent on tagUpdatedChan")
	}

	// Verify tag was deleted
	tags := service.GetTags()
	if len(tags) != 0 {
		t.Errorf("expected 0 tags in service, got %d", len(tags))
	}
}

func TestManagerService_Process_DeleteTags_NotFound(t *testing.T) {
	tagManagerChan := make(chan Manager, 10)
	tagUpdatedChan := make(chan llrp.Tags, 10)
	isConnAlive := &atomic.Bool{}
	isConnAlive.Store(true)

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}

	// Try to delete non-existent tag
	deleteCmd := Manager{Action: DeleteTags, Tags: llrp.Tags{tag1}}
	service.Process(deleteCmd)

	select {
	case result := <-tagManagerChan:
		if len(result.Tags) != 0 {
			t.Errorf("expected 0 tags in result (not found), got %d", len(result.Tags))
		}
	default:
		t.Error("no response received on tagManagerChan")
	}

	// Should not send update for non-existent tag
	select {
	case <-tagUpdatedChan:
		t.Error("should not send update for non-existent tag")
	default:
		// Expected - no update
	}
}

func TestManagerService_Process_RetrieveTags(t *testing.T) {
	tagManagerChan := make(chan Manager, 10)
	tagUpdatedChan := make(chan llrp.Tags, 10)
	isConnAlive := &atomic.Bool{}

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	tag2, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101011"})
	if err != nil {
		t.Fatalf("failed to create tag2: %v", err)
	}

	service.SetTags(llrp.Tags{tag1, tag2})

	retrieveCmd := Manager{Action: RetrieveTags, Tags: llrp.Tags{}}
	service.Process(retrieveCmd)

	select {
	case result := <-tagManagerChan:
		if len(result.Tags) != 2 {
			t.Errorf("expected 2 tags in result, got %d", len(result.Tags))
		}
	default:
		t.Error("no response received on tagManagerChan")
	}
}

func TestManagerService_Process_NoUpdateWhenConnNotAlive(t *testing.T) {
	tagManagerChan := make(chan Manager, 10)
	tagUpdatedChan := make(chan llrp.Tags, 10)
	isConnAlive := &atomic.Bool{}
	isConnAlive.Store(false)

	service := NewManagerService(tagManagerChan, tagUpdatedChan, isConnAlive)

	tag1, err := llrp.NewTag(&llrp.TagRecord{PCBits: "3000", EPC: "001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010"})
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	cmd := Manager{Action: AddTags, Tags: llrp.Tags{tag1}}

	service.Process(cmd)
	<-tagManagerChan // consume response

	// Should not send update when connection is not alive
	select {
	case <-tagUpdatedChan:
		t.Error("should not send update when connection is not alive")
	default:
		// Expected - no update
	}
}
