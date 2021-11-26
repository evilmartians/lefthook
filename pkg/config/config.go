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
