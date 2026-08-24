package config

// Config is the on-disk settings ocrogram will persist after TUI setup.
type Config struct {
	ScreenshotDir string
	Enabled       bool
}

// Load reads ~/.config/ocrogram/config.toml.
// TODO: implement load/save.
func Load() (Config, error) {
	return Config{}, nil
}
