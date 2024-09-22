package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/BurntSushi/toml"
)

type Rule struct {
	Regexp    string `toml:"regexp"`
	Colors    string `toml:"colors"`
	Overwrite bool   `toml:"overwrite"`
	Priority  int    `toml:"priority"`
}

func loadRules(ruleFile string) (map[string]Rule, error) {
	var rules map[string]Rule

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if verbose {
			fmt.Println("Error getting home directory:", err)
		}
		homeDir = ""
	}

	rulesPaths := []string{
		os.Getenv("COLORIZE_RULES"),
		filepath.Join(homeDir, ".config/colorize/rules"),
		"/etc/colorize.d/",
	}

	for _, rulesDir := range rulesPaths {
		if rulesDir == "" {
			continue
		}

		ruleFilePath := path.Join(rulesDir, ruleFile)
		fmt.Println(ruleFilePath)

		if _, err := os.Stat(ruleFilePath); err == nil {
			if verbose {
				fmt.Println("Using rules file:", ruleFilePath)
			}

			content, err := os.ReadFile(ruleFilePath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading file:", err)
			}

			re := regexp.MustCompile(`^\s*"?\$schema"?\s*=\s*[^#\n]*\s*\n?`)
			cleanContent := re.ReplaceAll(content, nil)
			fmt.Println(string(cleanContent))

			_, err = toml.Decode(string(cleanContent), &rules)
			if err == nil {
				type RuleSliceItem struct {
					Key   string
					Value Rule
				}
				var ruleSlice []RuleSliceItem

				for key, value := range rules {
					ruleSlice = append(ruleSlice, RuleSliceItem{key, value})
				}

				sort.Slice(ruleSlice, func(i int, j int) bool {
					// First sort by Overwrite (descending)
					if ruleSlice[i].Value.Overwrite != ruleSlice[j].Value.Overwrite {
						return ruleSlice[i].Value.Overwrite
					}
					return ruleSlice[i].Value.Priority > ruleSlice[j].Value.Priority
				})

				rules = make(map[string]Rule)

				for _, sliceItem := range ruleSlice {
					rules[sliceItem.Key] = sliceItem.Value
				}

				return rules, nil
			} else {
				fmt.Fprintln(os.Stderr, "Can't load rules from path:", err)
			}

		}
	}

	return nil, fmt.Errorf("No rules found.")
}
