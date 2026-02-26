// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
)

var (
	projectFlag string
	jsonOutput  bool
)

var rootCmd = &cobra.Command{
	Use:   "apikeyscan",
	Short: "A tool to scan and list Google Cloud API keys",
	Long: `apikeyscan audits your Google Cloud project for API keys.
It surfaces keys that are completely unrestricted, allowing you to 
identify and secure potential vulnerabilities in your infrastructure.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&projectFlag, "project", "p", "", "Google Cloud Project ID")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results in deterministic JSON format (Agent-friendly)")
}

func getProjectID() (string, error) {
	// 1. Check flag
	if projectFlag != "" {
		return projectFlag, nil
	}

	// 2. Check existing environment variable
	if p := os.Getenv("GOOGLE_CLOUD_PROJECT"); p != "" {
		return p, nil
	}

	// 3. Check ~/.env
	if homeDir, err := os.UserHomeDir(); err == nil {
		envPath := filepath.Join(homeDir, ".env")
		// We ignore the error; if the file doesn't exist, we just move on.
		_ = godotenv.Load(envPath)
	}

	// Re-check environment variable after potentially loading ~/.env
	if p := os.Getenv("GOOGLE_CLOUD_PROJECT"); p != "" {
		return p, nil
	}

	// 4. Fallback to ADC (Application Default Credentials)
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", err
	}
	if creds.ProjectID == "" {
		return "", fmt.Errorf("could not determine project ID from credentials. Please specify explicitly with --project flag, GOOGLE_CLOUD_PROJECT env var, or in ~/.env")
	}
	return creds.ProjectID, nil
}
