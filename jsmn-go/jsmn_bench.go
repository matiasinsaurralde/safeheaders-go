package jsmngo

import (
	"testing"
)

func BenchmarkParse(b *testing.B) {
	json := []byte(`{"key": "value", "arr": [1, 2, 3]}`) // Or load a large file.
	p := NewParser(10)
	for i := 0; i < b.N; i++ {
		p.Parse(json)
	}
}

func BenchmarkParseParallel(b *testing.B) {
	json := []byte(`{"key": "value", "arr": [1, 2, 3]}`) // Use large for real gains.
	for i := 0; i < b.N; i++ {
		ParseParallel(json, 10)
	}
}
