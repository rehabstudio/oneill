package oneill

import (
	"fmt"
	"os"
)

// ExitOnError checks that an error is not nil. If the passed value is an
// error, it is logged and the program exits with an error code of 1
func ExitOnError(err error, prefix string) {
	if err != nil {
		LogFatal(fmt.Sprintf("%s: %s", prefix, err))
		os.Exit(1)
	}
}
