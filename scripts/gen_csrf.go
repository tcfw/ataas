package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	d := make([]byte, 16)
	_, err := rand.Read(d)
	if err != nil {
		panic(err)
	}

	str := hex.EncodeToString(d)

	fmt.Printf("CSRF: %s\n", str)
}
