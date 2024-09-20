package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	ruleDir  string
	verbose  bool
	useColor bool
)

type Rule struct {
	Regexp string `toml:"regexp"`
	Colors string `toml:"colors"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Example: "colorize -- stat go.mod",
	Use:     "colorize [OPTOINS] -- COMMAND [OPTIONS/ARGUMENTS]",
	Short:   "A colorizer for your favorite commands",
	Run: func(cmd *cobra.Command, args []string) {
		useColor = true

		if len(args) < 1 {
			cmd.Help()
			os.Exit(0)
		}

		cmdName := args[0]
		cmdArgs := args[1:]

		config, err := loadConfig(verbose)

		if err != nil && verbose {
			fmt.Fprintln(os.Stderr, "Failed to load config:", err)
		}

		var ruleFileName string
		for name, values := range config {
			if cmdName == name {
				ruleFileName = values.File
			}

			commandStr := strings.Join(args, " ")
			if matched, _ := regexp.Match(values.Regexp, []byte(commandStr)); matched {
				ruleFileName = values.File
			}
		}

		var rules map[string]Rule

		if useColor && len(ruleFileName) > 0 {
			path := filepath.Join("rules", ruleFileName)

			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading file:", err)
			}

			re := regexp.MustCompile(`^\s*"?\$schema"?\s*=\s*[^#\n]*\s*\n?`)
			cleanContent := re.ReplaceAll(content, nil)

			_, err = toml.Decode(string(cleanContent), &rules)
			if verbose && err != nil {
				fmt.Fprintln(os.Stderr, "Can't load rules from path:", err)
			}
		}

		if verbose {
			fmt.Printf("%d rules found.\n", len(rules))
		}

		runCmd := exec.Command(cmdName, cmdArgs...)

		stdout, err := runCmd.StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating stdout pipe:", err)
			os.Exit(1)
		}
		if err := runCmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "Error starting command:", err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			srcLine := scanner.Text()

			coloredLine := ""
			if useColor && len(rules) > 0 {
				coloredLine = ColorizeLine(srcLine, rules)
			}

			if len(coloredLine) > 0 {
				fmt.Println(coloredLine)
			} else {
				fmt.Println(srcLine)
			}

		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading stdout: %s\n", err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
