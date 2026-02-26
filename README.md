# apikeyscan

`apikeyscan` is a standalone Go command-line tool that lists all Google Cloud API keys within a specified project. 

For security and ease of review, it defaults to a `list` view that outputs a clean table containing the key's display name, creation date, its **restriction status**, and a truncated view of the actual API key string.

**Crucially, unrestricted API keys are sorted to the top of the list so they can be immediately identified.**

This tool operates natively using the Google Cloud Go SDK, completely independent of `gcloud` during execution.

## 🤖 Agent-First Interoperability

`apikeyscan` is designed to be easily usable by CI/CD pipelines, DevOps tooling, and automated AI agents. 

*   **Deterministic Output:** Use the `--json` global flag to output `list` and `details` commands as structured, parseable JSON arrays/objects. 
*   **Color Degradation:** The standard table outputs use semantic coloring to highlight restricted vs. unrestricted keys. This safely degrades if piped to a file or if the `NO_COLOR` environment variable is set.

## Prerequisites

1. **Google Cloud Credentials:** The tool uses Application Default Credentials (ADC) to authenticate. If you haven't authenticated recently, you can log in via:
   ```bash
   gcloud auth application-default login
   ```
2. **API Enabled:** The target Google Cloud Project must have the API Keys API enabled.
   ```bash
   gcloud services enable apikeys.googleapis.com --project=YOUR_PROJECT_ID
   ```

## Installation

### Option 1: Quick Install (curl | bash)
We provide a convenience script that automatically downloads and installs the latest pre-compiled binary for your OS and Architecture to `/usr/local/bin`.

```bash
curl -sL https://raw.githubusercontent.com/ghchinoy/apikeyscan/main/scripts/install.sh | bash
```

### Option 2: Go Install
If you have Go 1.20+ installed, you can build and install the latest version directly to your `$GOPATH/bin`:

```bash
go install github.com/ghchinoy/apikeyscan@latest
```

### Option 3: Build from Source
Clone the repository and build the binary manually:

```bash
git clone https://github.com/ghchinoy/apikeyscan.git
cd apikeyscan
go build -o apikeyscan
```

## Usage

The tool determines which Google Cloud Project to query by checking the following in order of precedence:

1. The `--project` (or `-p`) command-line flag.
2. The `GOOGLE_CLOUD_PROJECT` environment variable in your active shell.
3. A `GOOGLE_CLOUD_PROJECT` variable defined in a `~/.env` file in your home directory.
4. The default project tied to your Application Default Credentials (ADC).

### Tip: Using Environment Variables (Recommended)

If you have the `gcloud` CLI installed and a project configured, you can easily export it to your environment so `apikeyscan` automatically picks it up for all subcommands:

```bash
# Set the environment variable using your currently active gcloud project
export GOOGLE_CLOUD_PROJECT=$(gcloud config get project)

# Or, add it to a .env file in your home directory
echo "GOOGLE_CLOUD_PROJECT=$(gcloud config get project)" >> ~/.env
```

## Commands

### 1. `list` (Default)
Lists all API keys, placing the unrestricted ones at the very top. It indicates how many APIs the key is restricted to, or prints the API name if it's only restricted to one.

```bash
# Human-readable table
apikeyscan list

# Machine-readable JSON
apikeyscan list --json | jq '.'
```

**Example Output:**
```text
Fetching API Keys for project: your-project-id

Display Name                        | Created      | API Restrictions               | API Key
---------------------------------------------------------------------------------------------------------
Developer Key (Testing)             | 2023-11-12   | Unrestricted (ALL APIs)        | AIzaSy...xZ19A
Internal Server Key                 | 2023-10-05   | Restricted (5 APIs)            | AIzaSy...pQ82B
Maps Integration Key                | 2024-01-20   | maps-backend.googleapis.com    | AIzaSy...mN34C

Total API keys found: 3
```

### 2. `details`
Fetches the deep details of a specific API key, including all specific `App Restrictions` (like IP addresses or iOS bundle IDs) and all explicit API targets.

You can provide just the UUID string or the full API key resource name. By default, the API Key string is truncated to prevent accidental exposure in terminal logs. Use `--full` to override.

```bash
# Human-readable colored output
apikeyscan details 1234abcd-5678-efgh-9012-ijklmnop

# View the full unmasked API Key string
apikeyscan details my-api-key --full

# Machine-readable JSON
apikeyscan details 1234abcd-5678-efgh-9012-ijklmnop --json | jq '.serverIps'
```

**Example Output:**
```text
=== API Key Details ===
Name:         projects/your-project-id/locations/global/keys/1234abcd-5678-efgh-9012-ijklmnop
Display Name: Internal Server Key
UID:          1234abcd-5678-efgh-9012-ijklmnop
Key String:   AIzaSyA...
Created:      2023-10-05 14:48:00 +0000 UTC

--- Restrictions ---
API Restrictions: Restricted to 5 API(s):
  - compute.googleapis.com
  - storage.googleapis.com
...
App Restrictions: Server (IP Addresses)
  - 192.168.1.100
  - 10.0.0.5
```

# License

Apache 2.0; see [`LICENSE`](LICENSE) for details.

# Disclaimer

This project is not an official Google project. It is not supported by
Google and Google specifically disclaims all warranties as to its quality,
merchantability, or fitness for a particular purpose.
