package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-tools/go-steputils/cache"
	"github.com/kballard/go-shellquote"
)

func exportEnvironmentWithEnvman(keyStr, valueStr string) error {
	cmd := command.New("envman", "add", "--key", keyStr)
	cmd.SetStdin(strings.NewReader(valueStr))
	return cmd.Run()
}

func failf(message string, args ...interface{}) {
	log.Errorf(message, args...)
	os.Exit(1)
}

func main() {
	// Collecting caches
	fmt.Println()
	log.Infof("Collecting gradle caches...")

	gradleCache := cache.New()
	homeDir := pathutil.UserHomeDir()

	includePths := []string{
		filepath.Join(homeDir, ".gradle"),
		filepath.Join(homeDir, ".kotlin"),
		filepath.Join(homeDir, ".android", "build-cache"),
	}
	excludePths := []string{
		"/*.lock",
		"/*.bin",
		"/*/build/*.json",
		"/*/build/*.xml",
		"/*/build/*.properties",
		"/*/build/*/zip-cache/*",
		"/*.log",
		"/*.txt",
		"/*.rawproto",
		"/*.ap_",
		"/*.apk",
	}

	projectRoot, err := filepath.Abs(filepath.Dir(configs.GradlewPath))
	if err != nil {
		log.Warnf("Cache collection skipped: failed to determine project root path.")
	} else {
		if err := filepath.Walk(projectRoot, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				if f.Name() == "build" {
					includePths = append(includePths, path)
				}
				if f.Name() == ".gradle" {
					includePths = append(includePths, path)
				}
			}
			return nil
		}); err != nil {
			log.Warnf("Cache collection skipped: failed to determine cache paths.")
		} else {

			gradleCache.IncludePath(strings.Join(includePths, "\n"))
			gradleCache.ExcludePath(strings.Join(excludePths, "\n"))

			if err := gradleCache.Commit(); err != nil {
				log.Warnf("Cache collection skipped: failed to commit cache paths.")
			}
		}
	}
}
