package configurators

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func Int(i *int, name string) {
	if env, err := strconv.Atoi(os.Getenv(name)); err == nil {
		*i = env
	}
}

func Float(i *float64, name string) {
	if env, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
		*i = env
	}
}

func MegaInt(f *int, name string) {
	if env, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
		*f = int(env * 1000000)
	}
}

func String(s *string, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		*s = env
	}
}

func StringSlice(s *[]string, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")

		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}

		*s = parts

		return
	}

	*s = []string{}
}

func StringSliceFile(s *[]string, filepath string) error {
	if len(filepath) == 0 {
		return nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("Can't open file %s\n", filepath)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if str := scanner.Text(); len(str) != 0 && !strings.HasPrefix(str, "#") {
			*s = append(*s, str)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Failed to read presets file: %s", err)
	}

	return nil
}

func StringMap(m *map[string]string, name string) error {
	if env := os.Getenv(name); len(env) > 0 {
		mm := make(map[string]string)

		keyvalues := strings.Split(env, ";")

		for _, keyvalue := range keyvalues {
			parts := strings.SplitN(keyvalue, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("Invalid key/value: %s", keyvalue)
			}
			mm[parts[0]] = parts[1]
		}

		*m = mm
	}

	return nil
}

func Bool(b *bool, name string) {
	if env, err := strconv.ParseBool(os.Getenv(name)); err == nil {
		*b = env
	}
}

func HexSlice(b *[][]byte, name string) error {
	var err error

	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")

		keys := make([][]byte, len(parts))

		for i, part := range parts {
			if keys[i], err = hex.DecodeString(part); err != nil {
				return fmt.Errorf("%s expected to be hex-encoded strings. Invalid: %s\n", name, part)
			}
		}

		*b = keys
	}

	return nil
}

func HexSliceFile(b *[][]byte, filepath string) error {
	if len(filepath) == 0 {
		return nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("Can't open file %s\n", filepath)
	}

	keys := [][]byte{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		part := scanner.Text()

		if len(part) == 0 {
			continue
		}

		if key, err := hex.DecodeString(part); err == nil {
			keys = append(keys, key)
		} else {
			return fmt.Errorf("%s expected to contain hex-encoded strings. Invalid: %s\n", filepath, part)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Failed to read file %s: %s", filepath, err)
	}

	*b = keys

	return nil
}

func Patterns(s *[]*regexp.Regexp, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")
		result := make([]*regexp.Regexp, len(parts))

		for i, p := range parts {
			result[i] = RegexpFromPattern(strings.TrimSpace(p))
		}

		*s = result
	} else {
		*s = []*regexp.Regexp{}
	}
}

func RegexpFromPattern(pattern string) *regexp.Regexp {
	var result strings.Builder
	// Perform prefix matching
	result.WriteString("^")
	for i, part := range strings.Split(pattern, "*") {
		// Add a regexp match all without slashes for each wildcard character
		if i > 0 {
			result.WriteString("[^/]*")
		}

		// Quote other parts of the pattern
		result.WriteString(regexp.QuoteMeta(part))
	}
	// It is safe to use regexp.MustCompile since the expression is always valid
	return regexp.MustCompile(result.String())
}
