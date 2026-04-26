package app

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/templates"
	"github.com/spf13/afero"
)

const (
	timestampBase    = 10
	timestampBitsize = 64

	configFileMode   = 0o666
	checksumFileMode = 0o644
)

var ErrNoConfigFile = errors.New("no lefthook config")
var lefthookChecksumRegexp = regexp.MustCompile(`(\w+)\s+(\d+)(?:\s+([\w,-]+))?`)

// ConfigService works with config.
type ConfigService struct {
	repo   *git.Repo
	logger *logger.Logger
	config *config.Config
}

func (s *ConfigService) Exists() bool {
	_, err := s.MainPath()
	return err == nil
}

func (s *ConfigService) MainPath() (string, error) {
	configOverride := os.Getenv("LEFTHOOK_CONFIG")
	if len(configOverride) != 0 {
		if !filepath.IsAbs(configOverride) {
			configOverride = filepath.Join(s.repo.RootPath, configOverride)
		}
		if ok, _ := afero.Exists(s.repo.Fs, configOverride); !ok {
			return "", fmt.Errorf("couldn't find config from LEFTHOOK_CONFIG: %s", configOverride)
		}
		return configOverride, nil
	}

	for _, name := range config.MainConfigNames {
		for _, extension := range config.Extensions {
			configPath := filepath.Join(s.repo.RootPath, name+extension)
			if ok, _ := afero.Exists(s.repo.Fs, configPath); ok {
				return configPath, nil
			}
		}
	}

	return "", ErrNoConfigFile
}

func (s *ConfigService) Load() (*config.Config, error) {
	if s.config != nil {
		return s.config, nil
	}

	newConfig, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	s.config = newConfig

	return s.config, nil
}

func (s *ConfigService) Reload() (*config.Config, error) {
	s.config = nil
	return s.Load()
}

func (s *ConfigService) Create() error {
	filePath := filepath.Join(s.repo.RootPath, config.DefaultConfigName)

	s.logger.Info("Creating config:", filePath)
	err := afero.WriteFile(s.repo.Fs, filePath, templates.Config(), configFileMode)
	if err != nil {
		return err
	}

	return nil
}

func (s *ConfigService) SourceDirs() ([]string, error) {
	cfg, err := s.Load()
	if err != nil {
		return nil, err
	}

	sourceDirs := []string{
		filepath.Join(s.repo.RootPath, cfg.SourceDir),
		filepath.Join(s.repo.RootPath, cfg.SourceDirLocal),

		// Additional source dirs to support .config/
		filepath.Join(s.repo.RootPath, ".config", "lefthook"),
		filepath.Join(s.repo.RootPath, ".config", "lefthook-local"),
	}

	for _, remote := range cfg.Remotes {
		if remote.Configured() {
			// Append only source_dir, because source_dir_local doesn't make sense
			sourceDirs = append(
				sourceDirs,
				filepath.Join(
					s.repo.RemoteFolder(remote.GitURL, remote.Ref),
					cfg.SourceDir,
				),
			)
		}
	}

	return sourceDirs, nil
}

// Synchronized returns whether configured hooks are installed.
func (s *ConfigService) Synchronized() bool {
	// Check checksum in a checksum file
	checksumFilepath := filepath.Join(s.repo.InfoPath, config.ChecksumFileName)
	file, err := s.repo.Fs.Open(checksumFilepath)
	if err != nil {
		return false
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Warnf("Could not close %s: %s", file.Name(), cErr)
		}
	}()

	scanner := bufio.NewScanner(file)

	var storedChecksum string
	var storedTimestamp int64
	// TODO(mrexox): move to a separate func
	// var storedHooks []string

	// Checksum format:
	// <md5sum> <timestamp> <hook1,hook2,hook3>
	for scanner.Scan() {
		match := lefthookChecksumRegexp.FindStringSubmatch(scanner.Text())
		if match != nil {
			storedChecksum = match[1]
			storedTimestamp, err = strconv.ParseInt(match[2], timestampBase, timestampBitsize)
			if err != nil {
				return false
			}
			// if len(match[3]) > 0 {
			// 	storedHooks = strings.Split(match[3], ",")
			// }

			break
		}
	}
	if err = scanner.Err(); err != nil {
		log.Warnf("Could not read %s: %s", file.Name(), err)
		return false
	}

	if len(storedChecksum) == 0 {
		return false
	}

	configTimestamp, err := s.lastUpdated()
	if err != nil {
		return false
	}

	if storedTimestamp == configTimestamp {
		return true
	}

	cfg, err := s.Load()
	if err != nil {
		return false
	}
	configChecksum, err := cfg.Md5()
	if err != nil {
		return false
	}

	return storedChecksum == configChecksum
}

func (s *ConfigService) ForEachRemote(f func(remote *config.Remote)) {
	cfg, err := s.Load()
	if err != nil {
		panic(err.Error())
	}

	for _, remote := range cfg.Remotes {
		f(remote)
	}
}

func (s *ConfigService) checksumFilepath() string {
	return filepath.Join(s.repo.InfoPath, config.ChecksumFileName)
}

func (s *ConfigService) lastUpdated() (int64, error) {
	configPath, err := s.MainPath()
	if err != nil {
		return 0, err
	}

	configFile, err := s.repo.Fs.Stat(configPath)
	if err != nil {
		return 0, err
	}

	return configFile.ModTime().Unix(), nil
}

func (s *ConfigService) loadConfig() (*config.Config, error) {
	cfg, err := config.Load(s.repo)

	// Reset loaded colors
	s.setColors(cfg.Colors)

	return cfg, err
}

func (s *ConfigService) setColors(colors any) {
	if colors == nil {
		return
	}

	switch colorsTyped := colors.(type) {
	case string:
		switch colorsTyped {
		case "on":
			s.logger.SetColors(logger.DefaultColors)
		case "off":
			s.logger.SetColors(logger.NoColors)
		default:
		}
	case bool:
		if colorsTyped {
			s.logger.SetColors(logger.DefaultColors)
		} else {
			s.logger.SetColors(logger.NoColors)
		}
	case map[string]any:
		s.logger.SetColors(map[logger.Color]color.Color{
			logger.ColorCyan:   lipgloss.Color(colorsTyped["cyan"].(string)),
			logger.ColorGray:   lipgloss.Color(colorsTyped["gray"].(string)),
			logger.ColorGreen:  lipgloss.Color(colorsTyped["green"].(string)),
			logger.ColorRed:    lipgloss.Color(colorsTyped["red"].(string)),
			logger.ColorYellow: lipgloss.Color(colorsTyped["yellow"].(string)),
		})
	default:
	}
}
