package feature

import (
	"strings"

	"winsbygroup.com/regserver/internal/featurevalue"
)

func ToInt(s string) int {
	switch strings.ToLower(s) {
	case "integer":
		return 0
	case "string":
		return 1
	case "values":
		return 2
	default:
		return -1
	}
}

// MergeWithOverrides returns a map of feature names to values, applying customer
// overrides where they exist and using defaults otherwise.
func MergeWithOverrides(defs []Feature, vals []featurevalue.FeatureValue) map[string]any {
	out := make(map[string]any)

	// Build lookup for customer overrides
	overrides := make(map[int64]string)
	for _, v := range vals {
		overrides[v.FeatureID] = v.FeatureValue
	}

	for _, d := range defs {
		if v, ok := overrides[d.FeatureID]; ok {
			out[d.FeatureName] = v
		} else {
			out[d.FeatureName] = d.DefaultValue
		}
	}

	return out
}
