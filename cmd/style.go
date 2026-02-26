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

import "github.com/charmbracelet/lipgloss"

var (
	// StyleAccent: Landmarks (Headers, Group Titles)
	StyleAccent = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#399ee6", Dark: "#59c2ff"})

	// StyleCommand: Scan Targets (Command names, Flags)
	StyleCommand = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#5c6166", Dark: "#bfbdb6"})

	// StylePass: Success states (Completed tasks, safe configurations)
	StylePass = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#86b300", Dark: "#c2d94c"})

	// StyleWarn: Warnings (Pending states, potentially unsafe configs)
	StyleWarn = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#f2ae49", Dark: "#ffb454"})

	// StyleFail: Error (Failed tasks, explicitly dangerous states like unrestricted APIs)
	StyleFail = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#f07171", Dark: "#f07178"})

	// StyleMuted: De-emphasis (Metadata, Types, Defaults, truncated strings)
	StyleMuted = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#828c99", Dark: "#6c7680"})

	// StyleID: Identifiers (UIDs)
	StyleID = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#46ba94", Dark: "#95e6cb"})
)
