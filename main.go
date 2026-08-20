package main

import (
	"fmt"

	"github.com/josh-baltar/forgepool/internal/dispatch"
)

func main() {
	p := dispatch.NewPool(1, 2, 3)
	fmt.Printf("forgepool: %d warm daemon connections ready\n", p.Size())
}
