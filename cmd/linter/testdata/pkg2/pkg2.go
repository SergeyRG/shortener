package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("this is okay inside main")
	os.Exit(0)
}

func test() {
	log.Fatal("this is okay inside main") // want "используется log.Fatal и/или os.Exit вне функции main пакета main"
	os.Exit(0)                            // want "используется log.Fatal и/или os.Exit вне функции main пакета main"
}
