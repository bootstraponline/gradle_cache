package main

import (
        "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/cache"
)

// ConfigsModel ...
type ConfigsModel struct {
	// Gradle Inputs
	GradlewPath              string
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		GradlewPath:              os.Getenv("gradlew_path"),
	}
}

func (configs ConfigsModel) print() {

	log.Infof("Configs:")
	log.Printf("- GradlewPath: %s", configs.GradlewPath)
}

func (configs ConfigsModel) validate() (string, error) {
	if configs.GradlewPath == "" {
		explanation := `
Using a Gradle Wrapper (gradlew) is required, as the wrapper is what makes sure
that the right Gradle version is installed and used for the build.
You can find more information about the Gradle Wrapper (gradlew),
and about how you can generate one (if you would not have one already
in the official guide at: https://docs.gradle.org/current/userguide/gradle_wrapper.html`

		return explanation, errors.New("no GradlewPath parameter specified")
	}
	if exist, err := pathutil.IsPathExists(configs.GradlewPath); err != nil {
		return "", fmt.Errorf("Failed to check if GradlewPath exist at: %s, error: %s", configs.GradlewPath, err)
	} else if !exist {
		return "", fmt.Errorf("GradlewPath not exist at: %s", configs.GradlewPath)
	}

	return "", nil
}


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
