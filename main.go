package main

import (
	"fmt"
	"log"
	"os"
	"rsc.io/quote"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
	fmt.Println(quote.Hello())
}
