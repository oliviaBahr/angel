package launchctl

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// LaunchctlData represents parsed launchctl print output
type LaunchctlData struct {
	data map[string]interface{}
}

// Get returns a string value for the given key, or empty string if not found
func (l *LaunchctlData) Get(key string) string {
	if val, exists := l.data[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
		// Convert other types to string
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// GetInterface returns the raw interface{} value for the given key, or nil if not found
func (l *LaunchctlData) GetInterface(key string) interface{} {
	if val, exists := l.data[key]; exists {
		return val
	}
	return nil
}

func parseLaunchctlPrint(input []byte) (*LaunchctlData, error) {
	lines := strings.Split(string(input), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// Parse service name line (e.g., "system/com.friday.kmonad = {")
	serviceLine := lines[0]
	if !strings.Contains(serviceLine, " = {") {
		return nil, fmt.Errorf("invalid service line: %s", serviceLine)
	}

	parts := strings.Split(serviceLine, " = {")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid service line format: %s", serviceLine)
	}

	// Parse the object content
	objLines := lines[1:]
	data, err := parseObject(objLines)
	if err != nil {
		return nil, err
	}

	return &LaunchctlData{data: data}, nil
}

func parseObject(lines []string) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if line == "" {
			i++
			continue
		}

		// Check for closing brace
		if line == "}" {
			break
		}

		// Parse key-value pair
		if strings.Contains(line, " = ") {
			key, value, err := parseKeyValue(line)
			if err != nil {
				return nil, err
			}

			// Check if value is an object/array
			if strings.TrimSpace(value) == "{" {
				// Find the matching closing brace
				braceCount := 1
				startPos := i + 1
				endPos := startPos

				// Continue from the next line to find the closing brace
				for j, line := range lines[startPos:] {
					line = strings.TrimSpace(line)
					endPos = startPos + j + 1

					switch line {
					case "{":
						braceCount++
					case "}":
						braceCount--
					}

					if braceCount == 0 {
						break
					}
				}

				// Parse the nested structure
				nestedLines := lines[startPos:endPos]
				nestedValue, err := parseNestedValue(nestedLines)
				if err != nil {
					return nil, err
				}

				obj[key] = nestedValue

				// Skip ahead to after the nested structure
				i = endPos
			} else {
				obj[key] = parseSimpleValue(value)
				i++
			}
		} else {
			i++
		}
	}

	return obj, nil
}

func parseKeyValue(line string) (string, string, error) {
	parts := strings.SplitN(line, " = ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid key-value line: %s", line)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func parseNestedValue(lines []string) (interface{}, error) {
	if len(lines) == 0 {
		return []string{}, nil
	}

	// Check if it's a map (contains =>)
	hasArrow := false
	for _, line := range lines {
		if strings.Contains(line, " => ") {
			hasArrow = true
			break
		}
	}

	if hasArrow {
		return parseMap(lines)
	}

	// Check if it's an object (contains =)
	hasEquals := false
	for _, line := range lines {
		if strings.Contains(line, " = ") {
			hasEquals = true
			break
		}
	}

	if hasEquals {
		return parseObject(lines)
	}

	// Otherwise it's a string array
	return parseStringArray(lines)
}

func parseMap(lines []string) (map[string]string, error) {
	m := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, " => ") {
			parts := strings.SplitN(line, " => ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				m[key] = value
			}
		}
	}

	return m, nil
}

// parseStringArray parses lines into a string array
func parseStringArray(lines []string) ([]string, error) {
	var items []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}

	return items, nil
}

func parseSimpleValue(value string) interface{} {
	// Check for pipe-separated values
	if strings.Contains(value, " | ") {
		parts := strings.Split(value, " | ")
		var result []string
		for _, part := range parts {
			result = append(result, strings.TrimSpace(part))
		}
		return result
	}

	// Check if it's a number
	if isNumber(value) {
		// Handle hex numbers
		if strings.HasPrefix(value, "0x") {
			val, err := strconv.ParseInt(value, 0, 64)
			if err != nil {
				return value // fallback to string
			}
			return int(val)
		}
		if num, err := strconv.Atoi(value); err == nil {
			return num
		}
	}

	return value
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}

	// Check for hex
	if strings.HasPrefix(s, "0x") {
		for _, r := range s[2:] {
			if !unicode.IsDigit(r) && !(r >= 'a' && r <= 'f') && !(r >= 'A' && r <= 'F') {
				return false
			}
		}
		return true
	}

	// Check for regular numbers
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
