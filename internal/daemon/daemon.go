package daemon

import "fmt"

// Run is the long-lived process launchd will start.
// TODO: watch the screenshot folder, call ocrogram-helper, write text to the clipboard.
func Run() error {
	fmt.Println("ocrogram daemon: not implemented yet")
	return nil
}
