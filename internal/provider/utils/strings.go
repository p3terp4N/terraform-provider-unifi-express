package utils

import (
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func IsStringValueNotEmpty(s basetypes.StringValue) bool {
	return !s.IsUnknown() && !s.IsNull() && s.ValueString() != ""
}

// JoinNonEmpty joins non-empty strings from a slice with the specified separator.
func JoinNonEmpty(elements []string, separator string) string {
	var nonEmpty []string
	for _, elem := range elements {
		if elem != "" {
			nonEmpty = append(nonEmpty, elem)
		}
	}
	return strings.Join(nonEmpty, separator)
}

// SplitAndTrim splits a string by the specified separator and trims whitespace from each element.
func SplitAndTrim(s string, separator string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, separator)
	var result []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func RemoveElements[S ~[]E, E comparable](first S, second S) S {
	var result S
	for _, category := range first {
		if !slices.Contains(second, category) {
			result = append(result, category)
		}
	}
	return result
}
