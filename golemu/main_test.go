package main

import (
//"reflect"
//"testing"

//"golang.org/x/net/websocket"
// for benchmark
//"github.com/iomz/go-llrp"
)

/*
// Benchmark LLRP frame construction
var benchTagString = "12288,18,96,302DB319A000004000000003\n16802,22,128,c4a301c70d36cb32920b1d31c2dc3482\n10665,16,80,dc20420c4c72cf4d76de\n"
var benchTags = loadTagsFromCSV(strings.Repeat(benchTagString, 50000))

func benchmarkLLRPFrame(max int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		var trds []*[]byte
		var tagCount = 0
		var trdIndex = 0
		for _, tag := range benchTags {
			tagCount++
			if tagCount > max {
				break
			} else {
				trd := buildTagReportDataParameter(tag)
				if len(trds) == 0 {
					trds = append(trds, &trd)
				} else {
					*(trds[trdIndex]) = append(*(trds[trdIndex]), trd...)
				}
			}
		}
		for _, trd := range trds {
			roar := llrp.ROAccessReport(*trd, messageID)
			_ = roar
		}
	}
}

func BenchmarkLLRPFrame1(b *testing.B)      { benchmarkLLRPFrame(1, b) }
func BenchmarkLLRPFrame10(b *testing.B)     { benchmarkLLRPFrame(10, b) }
func BenchmarkLLRPFrame100(b *testing.B)    { benchmarkLLRPFrame(100, b) }
func BenchmarkLLRPFrame1000(b *testing.B)   { benchmarkLLRPFrame(1000, b) }
func BenchmarkLLRPFrame10000(b *testing.B)  { benchmarkLLRPFrame(10000, b) }
func BenchmarkLLRPFrame100000(b *testing.B) { benchmarkLLRPFrame(100000, b) }
*/
