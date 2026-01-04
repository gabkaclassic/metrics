package fatal

import (
	"log"
)

func logFatalTest() {
	log.Fatal("Test message") // want "fatal log is not in main package"
}

func myLogFatalTest() {
	Fatal("...")
}

func Fatal(args ...any) {

}
