package internal

import (
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v4"
)

const preCommitConfigPath = ".pre-commit-config.yaml"

var schemaPreCommitHook = &PreCommitHook{
	ID:            "helm-values-schema",
	Name:          "Generate Helm values schema",
	Entry:         "helm values schema --git-add .",
	Language:      "system",
	Files:         "values\\.yaml$",
	PassFilenames: lo.ToPtr(false),
}
var docsPreCommitHook = &PreCommitHook{
	ID:            "helm-values-docs",
	Name:          "Generate Helm values documentation",
	Entry:         "helm values docs --git-add .",
	Language:      "system",
	Files:         "values\\.yaml$",
	PassFilenames: lo.ToPtr(false),
}

func newPreCommitConfig() *PreCommitConfig {
	return &PreCommitConfig{
		Repos: []*PreCommitRepo{},
	}
}

type PreCommitConfig struct {
	Repos []*PreCommitRepo `yaml:"repos"`
}

func (c *PreCommitConfig) addRepo(repo *PreCommitRepo) {
	c.Repos = append(c.Repos, repo)
}

func (c *PreCommitConfig) getRepo(name string) *PreCommitRepo {
	for i, repo := range c.Repos {
		if repo.Repo == name {
			return c.Repos[i]
		}
	}
	return nil
}

func newPreCommitRepo(name string, rev string) *PreCommitRepo {
	return &PreCommitRepo{
		Repo:  name,
		Rev:   rev,
		Hooks: []*PreCommitHook{},
	}
}

type PreCommitRepo struct {
	Repo  string           `yaml:"repo"`
	Rev   string           `yaml:"rev,omitempty"`
	Hooks []*PreCommitHook `yaml:"hooks"`
}

func (r *PreCommitRepo) setHooks(hooks []*PreCommitHook) {
	for _, hook := range hooks {
		r.setHook(hook)
	}
}

func (r *PreCommitRepo) setHook(hook *PreCommitHook) {
	for i, h := range r.Hooks {
		if h.ID == hook.ID {
			r.Hooks[i] = hook
			return
		}
	}
	r.Hooks = append(r.Hooks, hook)
}

type PreCommitHook struct {
	ID            string `yaml:"id"`
	Name          string `yaml:"name,omitempty"`
	Entry         string `yaml:"entry,omitempty"`
	Language      string `yaml:"language,omitempty"`
	Files         string `yaml:"files,omitempty"`
	PassFilenames *bool  `yaml:"pass_filenames,omitempty"`
}

func readPreCommitConfig(path string) (*PreCommitConfig, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to read config: %w", err)
	}

	var config PreCommitConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, false, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, true, nil
}

func writePreCommitConfig(path string, config *PreCommitConfig) error {
	output, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, output, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func InstallPreCommitHooks(logger *logrus.Logger) error {
	newHooks := []*PreCommitHook{
		schemaPreCommitHook,
		docsPreCommitHook,
	}

	config, exists, err := readPreCommitConfig(preCommitConfigPath)
	if err != nil {
		return err
	}
	if !exists {
		config = newPreCommitConfig()
	}

	localRepo := config.getRepo("local")
	if localRepo == nil {
		localRepo = newPreCommitRepo("local", "")
		config.addRepo(localRepo)
	}

	localRepo.setHooks(newHooks)

	if err := writePreCommitConfig(preCommitConfigPath, config); err != nil {
		return err
	}

	fmt.Println("Successfully installed pre-commit hooks. To activate the hooks, run:")
	fmt.Println("  pre-commit install")

	return nil
}
