package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configSourceDirKey      string      = "source_dir"
	configSourceDirLocalKey string      = "source_dir_local"
	rootExecutionRelPath    string      = "."
	configFileName          string      = "lefthook"
	configLocalFileName     string      = "lefthook-local"
	configExtension         string      = ".yml"
	defaultFilePermission   os.FileMode = 0755
	defaultDirPermission    os.FileMode = 0666
)

var (
	rootPath     string
	cfgFile      string
	originConfig *viper.Viper

	au aurora.Aurora
)

var rootCmd = &cobra.Command{
	Use:   "lefthook",
	Short: "CLI tool to manage Git hooks",
	Long: `After installation go to your project directory
and execute the following command:
lefthook install`,
}

func Execute() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			os.Exit(1)
		}
	}()

	err := rootCmd.Execute()
	check(err)
}

func init() {
	cobra.OnInitialize(initConfig)
	au = aurora.NewAurora(IsTTY())
	log.SetOutput(os.Stdout)
}

func initConfig() {
	log.SetFlags(0)

	setRootPath(rootExecutionRelPath)

	// store original config before merge
	originConfig = viper.New()
	originConfig.SetConfigName(configFileName)
	originConfig.AddConfigPath(rootExecutionRelPath)
	originConfig.ReadInConfig()

	viper.SetConfigName(configFileName)
	viper.AddConfigPath(rootExecutionRelPath)
	viper.SetDefault(configSourceDirKey, ".lefthook")
	viper.SetDefault(configSourceDirLocalKey, ".lefthook-local")
	viper.ReadInConfig()

	viper.SetConfigName(configLocalFileName)
	viper.MergeInConfig()

	viper.AutomaticEnv()
}

func getRootPath() string {
	return rootPath
}

func setRootPath(path string) {
	rootPath, _ = filepath.Abs(path)
}

func getSourceDir() string {
	return filepath.Join(getRootPath(), viper.GetString(configSourceDirKey))
}

func getLocalSourceDir() string {
	return filepath.Join(getRootPath(), viper.GetString(configSourceDirLocalKey))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// IsTTY returns true if program is running with TTY
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}
