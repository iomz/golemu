// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package golemu

// TagReportData holds an actual parameter in byte and
// how many tags are included in the parameter
type TagReportData struct {
	Parameter []byte
	TagCount  uint
}

// TagReportDataStack is a stack of TagReportData
type TagReportDataStack []*TagReportData

// TotalTagCounts returns how many tags are included in the TagReportDataStack
func (trds TagReportDataStack) TotalTagCounts() uint {
	ttc := uint(0)
	for _, trd := range trds {
		ttc += trd.TagCount
	}
	return ttc
}
