package main

import (
	"ks-ssg/internal/site"
	"log"
)

// main runs the site generator with the default configuration.
func main() {
	if err := site.Build(site.DefaultConfig()); err != nil {
		log.Fatal(err)
	}
}
