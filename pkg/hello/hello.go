package hello

import "fmt"

// Say returns a welcome message
func Say(thing string) string {
	return fmt.Sprintf("Hello %s", thing)
}
