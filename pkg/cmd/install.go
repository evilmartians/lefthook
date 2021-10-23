package cmd

import (
	"github.com/evilmartians/lefthook/pkg/config"
	"github.com/evilmartians/lefthook/pkg/git"
	"github.com/evilmartians/lefthook/pkg/log"
	"github.com/evilmartians/lefthook/pkg/templates"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	force      bool
	aggressive bool
)

func NewInstallCmd(rootCmd *cobra.Command) {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Write basic configuration file in your project repository. Or initialize existed config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installExecutor(appFs)
		},
	}

	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "reinstall hooks without checking config version")
	rootCmd.Flags().BoolVarP(&aggressive, "aggressive", "a", false, "remove all hooks from .git/hooks dir and install lefthook hooks")

	rootCmd.AddCommand(installCmd)
}

func installExecutor(fs afero.Fs) error {
	repo, err := git.NewRepository()
	if err != nil {
		return err
	}
	c, err := readOrCreateConfig(repo.RootPath(), fs)
	if err != nil {
		return err
	}

	return createHooks(c, repo, fs)
}

func readOrCreateConfig(path string, fs afero.Fs) (*config.Config, error) {
	log.Info("Searching config in:", path)
	if !configExists(path, fs) {
		log.Info("Config not found, creating...")
		if err := createConfig(path, fs); err != nil {
			return nil, err
		}
	}
	return config.Load(fs, path)
}

func configExists(path string, fs afero.Fs) bool {
	extensions := []string{".yml", ".yaml"}
	confPath := filepath.Join(path, "lefthook")
	for _, ext := range extensions {
		if result, _ := afero.Exists(fs, confPath+ext); result {
			return result
		}
	}
	return false
}

func createConfig(path string, fs afero.Fs) error {
	file := filepath.Join(path, "lefthook.yml")
	if err := afero.WriteFile(fs, file, templates.Config(), 0666); err != nil {
		return err
	}
	log.Println("Added config:", file)

	return nil
}

func createHooks(config *config.Config, repo *git.Repository, fs afero.Fs) error {
	//	if currentVersion || force || aggressive
	//		saveOldHooks
	//		deleteHooks
	//		createHooks (including default one)
	return nil
}
