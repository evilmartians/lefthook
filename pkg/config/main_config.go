package config

func NewMainConfig() *MainConfig {
	return &MainConfig{
		Colors:         true,
		SourceDir:      ".lefthook",
		SourceDirLocal: ".lefthook-local",
	}
}

type MainConfig struct {
	Colors         bool     `mapstructure:"colors"`
	Extends        []string `mapstructure:"extends"`
	MinVersion     string   `mapstructure:"min_version"`
	SkipOutput     []string `mapstructure:"skip_output"`
	SourceDir      string   `mapstructure:"source_dir"`
	SourceDirLocal string   `mapstructure:"source_dir_local"`
}
