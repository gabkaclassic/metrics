package exit

import (
	"os"
)

func osExitTest() {
	os.Exit(1) // want "call exit function is not in main package"
}

func myExitTest() {
	Exit(1)
}

func Exit(code int) {

}
