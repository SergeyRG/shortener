package pkg1

import (
	"log"
	"os"
)

func mulfunc(i int) (int, error) {
	return i * 2, nil
}

func errCheckFunc() {
	panic("test")     // want "используется встроенная функция panic"
	log.Fatal("test") // want "используется log.Fatal и/или os.Exit вне функции main пакета main"
	os.Exit(1)        // want "используется log.Fatal и/или os.Exit вне функции main пакета main"
}
