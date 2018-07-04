// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package golemu

// TagRecord stors the Tags contents in string with json tags
type TagRecord struct {
	PCBits string `json:"PCBits"`
	EPC    string `json:"EPC"`
}
