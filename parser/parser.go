package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xlab/c/cc"
)

type Config struct {
	Arch         string   `yaml:"Arch"`
	IncludePaths []string `yaml:"IncludePaths"`
	SourcesPaths []string `yaml:"SourcesPaths"`
	archBits     TargetArchBits
}

func NewConfig(paths ...string) *Config {
	return &Config{
		SourcesPaths: paths,
	}
}

func ParseWith(cfg *Config) (*cc.TranslationUnit, cc.DefinesMap, error) {
	if len(cfg.SourcesPaths) == 0 {
		return nil, nil, errors.New("parser: no target paths specified")
	}
	cfg, err := checkConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	predefined, ok := predefines[cfg.archBits]
	if !ok {
		predefined = predefinedBase
	}
	model := models[cfg.archBits]
	return cc.Parse(predefined, cfg.SourcesPaths, model, cc.SysIncludePaths(cfg.IncludePaths))
}

func checkConfig(cfg *Config) (*Config, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	if arch, ok := arches[cfg.Arch]; !ok {
		// default to 64-bit arch
		cfg.archBits = Arch64
	} else if arch != Arch32 && arch != Arch64 {
		// default to 64-bit arch
		cfg.archBits = Arch64
	}
	// workaround for cznic's cc (it panics if supplied path is a dir)
	var saneFiles []string
	for _, path := range cfg.SourcesPaths {
		if !filepath.IsAbs(path) {
			if hPath, err := findFile(path, cfg.IncludePaths); err != nil {
				err = fmt.Errorf("parser: file specified but not found: %s (include paths: %s)",
					path, strings.Join(cfg.IncludePaths, ", "))
				return nil, err
			} else {
				path = hPath
			}
		}
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			continue
		}
		if absPath, err := filepath.Abs(path); err != nil {
			path = absPath
		}
		saneFiles = append(saneFiles, path)
	}
	cfg.SourcesPaths = saneFiles
	return cfg, nil
}

func findFile(path string, includePaths []string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	for _, inc := range includePaths {
		result := filepath.Join(inc, path)
		if _, err := os.Stat(result); err == nil {
			return result, nil
		}
	}
	return "", errors.New("not found")
}