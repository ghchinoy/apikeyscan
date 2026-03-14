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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	apikeys "cloud.google.com/go/apikeys/apiv2"
	"cloud.google.com/go/apikeys/apiv2/apikeyspb"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

type APIKeySummary struct {
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	Uid              string `json:"uid"`
	CreateTime       string `json:"createTime"`
	RestrictionsText string `json:"restrictions"`
	IsUnrestricted   bool   `json:"isUnrestricted"`
	KeyTruncated     string `json:"keyTruncated"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys in the project",
	Long: `Scans your Google Cloud project for all API keys and lists them in a tabular format.
It highlights unrestricted keys so you can easily identify security risks.
In JSON mode, it provides a deterministic output suitable for parsing by automated agents or jq.`,
	Example: `  # List all keys using Application Default Credentials
  apikeyscan list

  # List keys for a specific project
  apikeyscan list --project=my-project-id

  # Output as JSON for agent/script parsing
  apikeyscan list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := getProjectID()
		if err != nil {
			return err
		}

		ctx := context.Background()
		client, err := apikeys.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create apikeys client: %v", err)
		}
		defer client.Close()

		parent := fmt.Sprintf("projects/%s/locations/global", projectID)
		req := &apikeyspb.ListKeysRequest{Parent: parent}

		var keys []*apikeyspb.Key
		it := client.ListKeys(ctx, req)
		for {
			key, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return formatAPIError(err, projectID)
			}
			keys = append(keys, key)
		}

		// Sort keys: Unrestricted first, then by creation date
		sort.Slice(keys, func(i, j int) bool {
			iUnrestricted := isUnrestrictedAPI(keys[i])
			jUnrestricted := isUnrestrictedAPI(keys[j])
			if iUnrestricted != jUnrestricted {
				return iUnrestricted // Unrestricted (true) comes before Restricted (false)
			}
			return keys[i].CreateTime.AsTime().Before(keys[j].CreateTime.AsTime())
		})

		var summaries []APIKeySummary

		for _, key := range keys {
			keyStringReq := &apikeyspb.GetKeyStringRequest{Name: key.Name}
			truncatedKey := "ERROR"
			resp, err := client.GetKeyString(ctx, keyStringReq)
			if err != nil {
				if !jsonOutput {
					log.Printf("Failed to get string for key %s: %v", key.Name, err)
				}
			} else {
				fullKey := resp.KeyString
				if len(fullKey) > 12 {
					truncatedKey = fullKey[:6] + "..." + fullKey[len(fullKey)-5:]
				} else {
					truncatedKey = fullKey
				}
			}

			displayName := key.DisplayName
			if displayName == "" {
				parts := strings.Split(key.Name, "/")
				displayName = parts[len(parts)-1]
			}

			summary := APIKeySummary{
				Name:             key.Name,
				DisplayName:      displayName,
				Uid:              key.Uid,
				CreateTime:       key.CreateTime.AsTime().Format(time.RFC3339),
				RestrictionsText: getRestrictionsText(key),
				IsUnrestricted:   isUnrestrictedAPI(key),
				KeyTruncated:     truncatedKey,
			}
			summaries = append(summaries, summary)
		}

		if jsonOutput {
			// Deterministic JSON output for Agents
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(summaries); err != nil {
				return fmt.Errorf("failed to encode json: %v", err)
			}
			return nil
		}

		// Standard CLI Output
		if len(keys) == 0 {
			fmt.Println("No API keys found in this project.")
			return nil
		}

		fmt.Println(StyleAccent.Render(fmt.Sprintf("Fetching API Keys for project: %s\n", projectID)))

		// Format headers by padding the raw strings FIRST, then coloring the padded strings
		h1 := StyleAccent.Render(fmt.Sprintf("%-35s", "Display Name"))
		h2 := StyleAccent.Render(fmt.Sprintf("%-12s", "Created"))
		h3 := StyleAccent.Render(fmt.Sprintf("%-30s", "API Restrictions"))
		h4 := StyleAccent.Render("API Key")

		fmt.Printf("%s | %s | %s | %s\n", h1, h2, h3, h4)
		fmt.Println(StyleCommand.Render(strings.Repeat("-", 105)))

		for _, s := range summaries {
			dispName := s.DisplayName
			if len(dispName) > 33 {
				dispName = dispName[:30] + "..."
			}
			createdFormatted := s.CreateTime[:10] // Just the date part

			// 1. Pad the raw strings to fixed widths
			paddedName := fmt.Sprintf("%-35s", dispName)
			paddedDate := fmt.Sprintf("%-12s", createdFormatted)
			paddedRest := fmt.Sprintf("%-30s", s.RestrictionsText)
			paddedKey := fmt.Sprintf("%-15s", s.KeyTruncated)

			// 2. Apply colors to the already-padded blocks
			coloredName := paddedName // We leave name uncolored/default
			coloredDate := StyleMuted.Render(paddedDate)
			coloredKey := StyleMuted.Render(paddedKey)

			var coloredRest string
			if s.IsUnrestricted {
				if s.RestrictionsText == "Unrestricted (ALL APIs)*" {
					coloredRest = StyleWarn.Render(paddedRest) // App-restricted but API open
				} else {
					coloredRest = StyleFail.Render(paddedRest) // Fully open
				}
			} else {
				coloredRest = StylePass.Render(paddedRest)
			}

			// 3. Print the blocks separated by a static delimiter
			fmt.Printf("%s | %s | %s | %s\n", coloredName, coloredDate, coloredRest, coloredKey)
		}

		fmt.Printf("\nTotal API keys found: %d\n", len(keys))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func isUnrestrictedAPI(k *apikeyspb.Key) bool {
	return k.Restrictions == nil || len(k.Restrictions.ApiTargets) == 0
}

func getRestrictionsText(k *apikeyspb.Key) string {
	if k.Restrictions == nil {
		return "Unrestricted (ALL APIs)"
	}
	apiTargets := k.Restrictions.ApiTargets
	if len(apiTargets) == 0 {
		return "Unrestricted (ALL APIs)*"
	}
	if len(apiTargets) == 1 {
		svc := apiTargets[0].Service
		if len(svc) > 28 {
			svc = svc[:25] + "..."
		}
		return svc
	}
	return fmt.Sprintf("Restricted (%d APIs)", len(apiTargets))
}
