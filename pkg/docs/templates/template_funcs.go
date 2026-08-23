package templates

import (
	"fmt"
	"strings"
)

func lpad(s string, padStr string, pLen int64) string {
	remaining := int(pLen) - len(s)
	if remaining <= 0 {
		return s
	}

	return strings.Repeat(padStr, remaining) + s
}

func rpad(s string, padStr string, pLen int64) string {
	remaining := int(pLen) - len(s)
	if remaining <= 0 {
		return s
	}

	return s + strings.Repeat(padStr, remaining)
}

func maxLen(items []string) int {
	longest := 0

	for _, s := range items {
		for line := range strings.SplitSeq(s, "\n") {
			if len(line) > longest {
				longest = len(line)
			}
		}
	}

	return longest
}

func rowSelect(items []ValuesRow, field string) []string {
	var result []string

	for _, item := range items {
		switch field {
		case "Key":
			result = append(result, item.Key)
		case "Type":
			result = append(result, item.Type)
		case "Default":
			result = append(result, item.Default)
		case "Description":
			result = append(result, item.Description)
		}
	}

	return result
}

func mdRow(cols []string, colWidths []int64) string {
	c := make([]string, 0, len(cols))
	for i, col := range cols {
		c = append(c, rpad(col, " ", colWidths[i]))
	}

	return fmt.Sprintf("| %s |", strings.Join(c, " | "))
}

func mdMultiline(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)

	return strings.ReplaceAll(s, "\n", "</br>")
}
