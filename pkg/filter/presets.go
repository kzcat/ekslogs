package filter

// UnifiedPresetFilter defines a filter template with pattern type information
type UnifiedPresetFilter struct {
	Description string
	LogTypes    []string
	Pattern     string
	PatternType string // "simple", "optional", "exclude", "json", "regex"
	Advanced    bool   // Whether this is an advanced pattern
}

// UnifiedPresets combines both basic and advanced presets
var UnifiedPresets = map[string]UnifiedPresetFilter{
	// Basic presets (simple patterns)
	"auth-failures": {
		Description: "Authentication failures",
		LogTypes:    []string{"authenticator", "api"},
		Pattern:     "unauthorized",
		PatternType: "simple",
		Advanced:    false,
	},
	"api-errors": {
		Description: "API server errors",
		LogTypes:    []string{"api"},
		Pattern:     "ERROR",
		PatternType: "simple",
		Advanced:    false,
	},
	"audit-privileged": {
		Description: "Privileged operations in audit logs",
		LogTypes:    []string{"audit"},
		Pattern:     "create",
		PatternType: "simple",
		Advanced:    false,
	},
	"scheduler-issues": {
		Description: "Scheduler issues",
		LogTypes:    []string{"scheduler"},
		Pattern:     "error",
		PatternType: "simple",
		Advanced:    false,
	},
	"controller-issues": {
		Description: "Controller manager issues",
		LogTypes:    []string{"kcm"},
		Pattern:     "error",
		PatternType: "simple",
		Advanced:    false,
	},
	"cloud-issues": {
		Description: "Cloud controller manager issues",
		LogTypes:    []string{"ccm"},
		Pattern:     "error",
		PatternType: "simple",
		Advanced:    false,
	},
	"resource-issues": {
		Description: "Resource exhaustion or limits",
		LogTypes:    []string{"api", "scheduler"},
		Pattern:     "limit",
		PatternType: "simple",
		Advanced:    false,
	},
	"network-issues": {
		Description: "Network related issues",
		LogTypes:    []string{"api", "kcm", "ccm"},
		Pattern:     "network",
		PatternType: "simple",
		Advanced:    false,
	},

	// Advanced presets (complex patterns)
	"auth-issues-adv": {
		Description: "Authentication and authorization issues (advanced)",
		LogTypes:    []string{"authenticator", "api"},
		Pattern:     "?unauthorized ?\"permission denied\" ?\"authentication failed\" ?\"access denied\" ?forbidden",
		PatternType: "optional",
		Advanced:    true,
	},
	"critical-api-errors": {
		Description: "Critical API server errors excluding common warnings",
		LogTypes:    []string{"api"},
		Pattern:     "ERROR CRITICAL -warning -\"deadline exceeded\"",
		PatternType: "exclude",
		Advanced:    true,
	},
	"privileged-admin-actions": {
		Description: "Privileged admin actions in audit logs",
		LogTypes:    []string{"audit"},
		Pattern:     "{ $.user.username = \"admin\" } { $.verb = \"delete\" }",
		PatternType: "json",
		Advanced:    true,
	},
	"pod-scheduling-failures": {
		Description: "Pod scheduling failures",
		LogTypes:    []string{"scheduler"},
		Pattern:     "\"failed to schedule pod\" \"insufficient resources\"",
		PatternType: "simple",
		Advanced:    true,
	},
	"controller-reconcile-errors": {
		Description: "Controller reconciliation errors",
		LogTypes:    []string{"kcm"},
		Pattern:     "%reconcile.*failed%",
		PatternType: "regex",
		Advanced:    true,
	},
	"memory-pressure": {
		Description: "Memory pressure and OOM events",
		LogTypes:    []string{"api", "kcm"},
		Pattern:     "OOM killed memory",
		PatternType: "simple",
		Advanced:    true,
	},
	"security-events": {
		Description: "Security related events",
		LogTypes:    []string{"api", "audit", "authenticator"},
		Pattern:     "?\"security breach\" ?\"unauthorized access\" ?\"suspicious activity\" ?\"token expired\" ?\"certificate expired\"",
		PatternType: "optional",
		Advanced:    true,
	},
	"network-timeouts": {
		Description: "Network timeout issues",
		LogTypes:    []string{"api", "kcm", "ccm"},
		Pattern:     "%timeout.*network|network.*timeout%",
		PatternType: "regex",
		Advanced:    true,
	},
}

// GetUnifiedPreset returns a preset filter by name
func GetUnifiedPreset(name string) (UnifiedPresetFilter, bool) {
	preset, exists := UnifiedPresets[name]
	return preset, exists
}

// ListUnifiedPresets returns all available preset names
func ListUnifiedPresets() []string {
	var names []string
	for name := range UnifiedPresets {
		names = append(names, name)
	}
	return names
}

// ListBasicPresets returns basic preset names
func ListBasicPresets() []string {
	var names []string
	for name, preset := range UnifiedPresets {
		if !preset.Advanced {
			names = append(names, name)
		}
	}
	return names
}

// ListAdvancedPresets returns advanced preset names
func ListAdvancedPresets() []string {
	var names []string
	for name, preset := range UnifiedPresets {
		if preset.Advanced {
			names = append(names, name)
		}
	}
	return names
}
