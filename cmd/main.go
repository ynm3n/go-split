package main

import (
	split "github.com/ynm3n/go-split"
)

func main() {
	cf, err := split.ParseConfig()
	if err != nil {
		panic(err)
	}

	if err := split.Do(cf); err != nil {
		panic(err)
	}
}
