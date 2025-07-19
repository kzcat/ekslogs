package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/kzcat/ekslogs/pkg/filter"
	"github.com/spf13/cobra"
)

var (
	showAdvanced bool
	showAll      bool
)

var unifiedPresetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List available filter presets",
	Long:  `List all available filter presets for common use cases.`,
	Run: func(cmd *cobra.Command, args []string) {
		var presetNames []string

		if showAll {
			presetNames = filter.ListUnifiedPresets()
		} else if showAdvanced {
			presetNames = filter.ListAdvancedPresets()
		} else {
			presetNames = filter.ListBasicPresets()
		}

		// Sort preset names for consistent output
		sort.Strings(presetNames)

		if showAdvanced {
			fmt.Println("Available advanced filter presets:")
		} else if showAll {
			fmt.Println("Available filter presets (basic and advanced):")
		} else {
			fmt.Println("Available basic filter presets:")
		}
		fmt.Println()

		for _, name := range presetNames {
			preset, _ := filter.GetUnifiedPreset(name)

			// Print preset name and description
			if preset.Advanced {
				color.New(color.FgMagenta, color.Bold).Printf("  %s\n", name)
			} else {
				color.New(color.FgCyan, color.Bold).Printf("  %s\n", name)
			}
			fmt.Printf("    Description: %s\n", preset.Description)
			fmt.Printf("    Log types: %s\n", strings.Join(preset.LogTypes, ", "))
			fmt.Printf("    Pattern: %s\n", preset.Pattern)

			if showAll || showAdvanced {
				fmt.Printf("    Pattern type: %s\n", preset.PatternType)
			}
			fmt.Println()
		}

		fmt.Println("Usage example:")
		fmt.Println("  ekslogs my-cluster -p api-errors")
		fmt.Println("  ekslogs my-cluster -p network-timeouts -F")
		fmt.Println()

		if showAll || showAdvanced {
			fmt.Println("Pattern types:")
			fmt.Println("  - simple: Multiple terms (AND condition)")
			fmt.Println("  - optional: Terms with '?' prefix (OR condition)")
			fmt.Println("  - exclude: Terms with '-' prefix are excluded")
			fmt.Println("  - json: JSON structure filtering")
			fmt.Println("  - regex: Regular expression pattern (enclosed in %)")
			fmt.Println()
		}

		if !showAdvanced && !showAll {
			fmt.Println("To see advanced presets, run: ekslogs presets --advanced")
			fmt.Println("To see all presets, run: ekslogs presets --all")
		}
	},
}

func init() {
	rootCmd.AddCommand(unifiedPresetsCmd)
	unifiedPresetsCmd.Flags().BoolVar(&showAdvanced, "advanced", false, "Show only advanced presets")
	unifiedPresetsCmd.Flags().BoolVar(&showAll, "all", false, "Show all presets (basic and advanced)")
}
