package service

import "fmt"

// Start installs the LaunchAgent and loads it so ocrogram runs at login.
// TODO: write contrib/com.joelpeckham.ocrogram.plist into ~/Library/LaunchAgents and launchctl load.
func Start() error {
	fmt.Println("ocrogram start: LaunchAgent install is not implemented yet")
	return nil
}

// Stop unloads and removes the LaunchAgent.
// TODO: launchctl unload and remove the plist.
func Stop() error {
	fmt.Println("ocrogram stop: LaunchAgent removal is not implemented yet")
	return nil
}
