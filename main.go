package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/cache"
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
		"~/.gradle/**",
		"~/.android/build-cache/**",
		"*.lock",
		"*.bin",
		"/**/build/**.json",
		"/**/build/**.xml",
		"/**/build/**.properties",
		"/**/build/**/zip-cache/**",
		"*.log",
		"*.txt",
		"*.rawproto",
		"!*.ap_",
		"!*.apk",
	}

	projectRoot, err := filepath.Abs(os.Getenv("BITRISE_SOURCE_DIR"))
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
