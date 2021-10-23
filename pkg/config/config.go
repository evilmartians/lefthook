package config

type Config struct {
	Colors         bool     `mapstructure:"colors"`
	Extends        []string `mapstructure:"extends"`
	MinVersion     string   `mapstructure:"min_version"`
	SkipOutput     []string `mapstructure:"skip_output"`
	SourceDir      string   `mapstructure:"source_dir"`
	SourceDirLocal string   `mapstructure:"source_dir_local"`

	Hooks map[string]*Hook
}

// Merge another config into current. Current is overwritten.
func (c *Config) Merge(another *Config) {
	if another == nil {
		return
	}

	c.Colors = another.Colors

	if len(another.Extends) != 0 {
		c.Extends = another.Extends
	}
	if another.MinVersion != "" {
		c.MinVersion = another.MinVersion
	}
	if len(another.SkipOutput) != 0 {
		c.SkipOutput = another.SkipOutput
	}
	if another.SourceDir != "" {
		c.SourceDir = another.SourceDir
	}
	if another.SourceDirLocal != "" {
		c.SourceDirLocal = another.SourceDirLocal
	}

	for key, anotherHook := range another.Hooks {
		if c.Hooks[key] == nil {
			c.Hooks[key] = anotherHook
		} else {
			c.Hooks[key].Merge(anotherHook)
		}
	}
}
