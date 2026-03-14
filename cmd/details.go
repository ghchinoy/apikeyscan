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
	"os"
	"strings"

	apikeys "cloud.google.com/go/apikeys/apiv2"
	"cloud.google.com/go/apikeys/apiv2/apikeyspb"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

var showFullKey bool

type APIKeyDetail struct {
	Name             string   `json:"name"`
	DisplayName      string   `json:"displayName"`
	Uid              string   `json:"uid"`
	KeyString        string   `json:"keyString"`
	CreateTime       string   `json:"createTime"`
	UpdateTime       string   `json:"updateTime,omitempty"`
	IsUnrestricted   bool     `json:"isUnrestricted"`
	APITargets       []string `json:"apiTargets,omitempty"`
	BrowserReferrers []string `json:"browserReferrers,omitempty"`
	ServerIps        []string `json:"serverIps,omitempty"`
	AndroidApps      []string `json:"androidApps,omitempty"`
	IosBundles       []string `json:"iosBundles,omitempty"`
}

var detailsCmd = &cobra.Command{
	Use:   "details [KEY_ID_OR_NAME]",
	Short: "Get details for a specific API key",
	Long: `Provides a deep-dive view into a specific API key's configuration.
It displays all assigned application restrictions (IPs, HTTP Referrers, iOS/Android apps) 
as well as explicit API targets.

By default, the API key string itself is truncated to prevent accidental exposure 
(e.g., in screenshots or CI/CD logs). Use the --full flag to display the entire key.

If the key is unrestricted, it will prominently warn the user.`,
	Example: `  # Get details with a safely truncated key string
  apikeyscan details my-api-key

  # Expose the full, unmasked API key string
  apikeyscan details my-api-key --full

  # Output full details as JSON (key will be truncated unless --full is passed)
  apikeyscan details my-api-key --json`,
	Args: cobra.ExactArgs(1),
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

		keyArg := args[0]
		var targetKey *apikeyspb.Key

		// 1. Try fetching it directly
		nameToFetch := keyArg
		if !strings.HasPrefix(nameToFetch, "projects/") {
			nameToFetch = fmt.Sprintf("projects/%s/locations/global/keys/%s", projectID, nameToFetch)
		}

		req := &apikeyspb.GetKeyRequest{Name: nameToFetch}
		targetKey, err = client.GetKey(ctx, req)

		// 2. Search by Display Name or UID if not found
		if err != nil {
			parent := fmt.Sprintf("projects/%s/locations/global", projectID)
			it := client.ListKeys(ctx, &apikeyspb.ListKeysRequest{Parent: parent})
			found := false

			for {
				k, errList := it.Next()
				if errList == iterator.Done {
					break
				}
				if errList != nil {
					return formatAPIError(errList, projectID)
				}

				if k.DisplayName == keyArg || k.Uid == keyArg {
					targetKey = k
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("could not find an API key matching ID, UID, or Display Name '%s'", keyArg)
			}
		}

		// 3. Fetch the string value
		keyStringReq := &apikeyspb.GetKeyStringRequest{Name: targetKey.Name}
		finalKeyString := "ERROR (could not fetch)"
		resp, err := client.GetKeyString(ctx, keyStringReq)
		if err == nil {
			rawKey := resp.KeyString
			// Truncate the key unless the --full flag is explicitly passed
			if !showFullKey && len(rawKey) > 12 {
				finalKeyString = rawKey[:6] + "..." + rawKey[len(rawKey)-5:]
			} else {
				finalKeyString = rawKey
			}
		}

		// 4. Build Detail Struct
		detail := APIKeyDetail{
			Name:           targetKey.Name,
			DisplayName:    targetKey.DisplayName,
			Uid:            targetKey.Uid,
			KeyString:      finalKeyString,
			CreateTime:     targetKey.CreateTime.AsTime().String(),
			IsUnrestricted: targetKey.Restrictions == nil || len(targetKey.Restrictions.ApiTargets) == 0,
		}
		if targetKey.UpdateTime != nil {
			detail.UpdateTime = targetKey.UpdateTime.AsTime().String()
		}

		if targetKey.Restrictions != nil {
			for _, t := range targetKey.Restrictions.ApiTargets {
				detail.APITargets = append(detail.APITargets, t.Service)
			}
			if browser := targetKey.Restrictions.GetBrowserKeyRestrictions(); browser != nil {
				detail.BrowserReferrers = browser.AllowedReferrers
			} else if server := targetKey.Restrictions.GetServerKeyRestrictions(); server != nil {
				detail.ServerIps = server.AllowedIps
			} else if android := targetKey.Restrictions.GetAndroidKeyRestrictions(); android != nil {
				for _, app := range android.AllowedApplications {
					detail.AndroidApps = append(detail.AndroidApps, fmt.Sprintf("%s (%s)", app.PackageName, app.Sha1Fingerprint))
				}
			} else if ios := targetKey.Restrictions.GetIosKeyRestrictions(); ios != nil {
				detail.IosBundles = ios.AllowedBundleIds
			}
		}

		if jsonOutput {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(detail); err != nil {
				return fmt.Errorf("failed to encode json: %v", err)
			}
			return nil
		}

		// 5. Standard CLI Output
		fmt.Println(StyleAccent.Render("=== API Key Details ==="))
		fmt.Printf("Name:         %s\n", StyleMuted.Render(detail.Name))
		fmt.Printf("Display Name: %s\n", detail.DisplayName)
		fmt.Printf("UID:          %s\n", StyleID.Render(detail.Uid))
		fmt.Printf("Key String:   %s\n", StyleMuted.Render(detail.KeyString))
		fmt.Printf("Created:      %s\n", StyleMuted.Render(detail.CreateTime))
		if detail.UpdateTime != "" {
			fmt.Printf("Updated:      %s\n", StyleMuted.Render(detail.UpdateTime))
		}

		fmt.Println(StyleAccent.Render("\n--- Restrictions ---"))

		if targetKey.Restrictions == nil {
			fmt.Println(StyleFail.Render("Status: FULLY UNRESTRICTED (Dangerous!)"))
			fmt.Println(StyleMuted.Render("- This key can call any API enabled in the project."))
			fmt.Println(StyleMuted.Render("- This key can be used from any IP address or application."))
		} else {
			// Print API restrictions
			if len(detail.APITargets) == 0 {
				fmt.Println("API Restrictions:", StyleWarn.Render("Unrestricted (Can call ANY enabled API)"))
			} else {
				fmt.Printf("API Restrictions: Restricted to %d API(s):\n", len(detail.APITargets))
				for _, target := range targetKey.Restrictions.ApiTargets {
					fmt.Printf("  - %s\n", StyleCommand.Render(target.Service))
					if len(target.Methods) > 0 {
						fmt.Printf("    Methods: %s\n", StyleMuted.Render(strings.Join(target.Methods, ", ")))
					}
				}
			}

			// Print Application restrictions
			fmt.Print("App Restrictions: ")

			if len(detail.BrowserReferrers) > 0 {
				fmt.Println("Browser (HTTP Referrers)")
				for _, r := range detail.BrowserReferrers {
					fmt.Printf("  - %s\n", StyleMuted.Render(r))
				}
			} else if len(detail.ServerIps) > 0 {
				fmt.Println("Server (IP Addresses)")
				for _, ip := range detail.ServerIps {
					fmt.Printf("  - %s\n", StyleMuted.Render(ip))
				}
			} else if len(detail.AndroidApps) > 0 {
				fmt.Println("Android Apps")
				for _, app := range detail.AndroidApps {
					fmt.Printf("  - %s\n", StyleMuted.Render(app))
				}
			} else if len(detail.IosBundles) > 0 {
				fmt.Println("iOS Apps")
				for _, bundle := range detail.IosBundles {
					fmt.Printf("  - %s\n", StyleMuted.Render(bundle))
				}
			} else {
				fmt.Println(StyleMuted.Render("None (Can be used from anywhere)"))
			}
		}

		if detail.IsUnrestricted {
			fmt.Println(StyleFail.Render("\nThis API key is unrestricted. To prevent unauthorized use and quota theft, restrict your key to limit how it can be used."))
		}

		return nil
	},
}

func init() {
	detailsCmd.Flags().BoolVar(&showFullKey, "full", false, "Show the full API key string instead of truncating it")
	rootCmd.AddCommand(detailsCmd)
}
