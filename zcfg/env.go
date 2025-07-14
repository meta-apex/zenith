package zcfg

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var (
	envOnce sync.Once
)

// loadEnvFile loads .env file if it exists
func loadEnvFile() {
	envOnce.Do(func() {
		if _, err := os.Stat(".env"); err == nil {
			_ = godotenv.Load(".env")
		}
	})
}

// processEnvVars processes environment variables in a string value
func processEnvVars(value string, useEnv bool) (string, error) {
	if !useEnv {
		// Check if value contains env var syntax
		if strings.Contains(value, "${") {
			return "", fmt.Errorf("environment variables not enabled but found env var syntax in value: %s", value)
		}
		return value, nil
	}

	// Regex to match ${VAR} and ${VAR:default}
	envVarRegex := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)

	// First, check for any environment variables without default values that don't exist
	matches := envVarRegex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			envVar := match[1]
			envValue := os.Getenv(envVar)

			// Check if there's a default value (match[2] exists and is not empty or colon is present)
			hasDefault := len(match) > 2 && strings.Contains(match[0], ":")

			// If env var doesn't exist and no default value provided
			if envValue == "" && !hasDefault {
				return "", fmt.Errorf("environment variable %s not found", envVar)
			}
		}
	}

	result := envVarRegex.ReplaceAllStringFunc(value, func(match string) string {
		submatches := envVarRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		envVar := submatches[1]
		defaultValue := ""
		if len(submatches) > 2 {
			defaultValue = submatches[2]
		}

		envValue := os.Getenv(envVar)
		if envValue != "" {
			return envValue
		}

		return defaultValue
	})

	return result, nil
}

// processEnvValue processes environment variables in any value
func processEnvValue(value any, useEnv bool) (any, error) {
	switch v := value.(type) {
	case string:
		return processEnvVars(v, useEnv)
	default:
		return value, nil
	}
}
