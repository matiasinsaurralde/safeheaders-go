# SafeHeaders-Go

A collection of idiomatic Go ports of popular single-header C libraries, enhanced with Go's concurrency and safety features. These ports eliminate C's raw pointer risks using Go's slices and bounds checking, while adding novel twists like parallel processing for high-throughput scenarios.

## Why?
- **Safety**: No buffer overflows or undefined behavior.
- **Performance**: Leverage goroutines for concurrency, e.g., parallel tokenizing.
- **Simplicity**: Drop-in packages for embedded, web, or edge apps.
- **Novelty**: Go-specific features like streaming I/O for real-time data (e.g., IoT JSON streams).

## Current Ports
- [jsmn-go](./jsmn-go): Lightweight JSON tokenizer with parallel and streaming support.

## Usage Example (jsmn-go)

```go
package main

import (
	"fmt"
	"github.com/alikatgh/safeheaders-go/jsmn-go"
)

func main() {
	json := []byte(`{"key": "value"}`)
	tokens, err := jsmngo.ParseParallel(json, 10)
	if err != nil {
		panic(err)
	}
	for _, tok := range tokens {
		fmt.Printf("Token: %v\n", tok)
	}
}
```

## Contributing
Pick a single-header C lib from the wishlist below, port it to pure Go, add one concurrent enhancement, and PR!
- Wishlist:
  - stb_image.h (images)
  - stb_truetype.h (fonts)
  - miniz.h (compression)
  - linenoise.h (CLI input).

Guidelines: Keep it allocation-free where possible, include benchmarks/tests, and benchmark vs. original C.

## License
MIT
