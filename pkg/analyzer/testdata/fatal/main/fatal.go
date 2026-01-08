package main

import (
	"log"
)

func logFatalTest() {
	log.Fatal("Test message")
}

func myLogFatalTest() {
	Fatal("...")
}

func Fatal(args ...any) {

}
