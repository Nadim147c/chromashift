package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	rulesDir string
	verbose  bool
	useColor bool
)

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

		config, err := loadConfig()

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
			rules, err = loadRules(ruleFileName)
			if verbose && err != nil {
				fmt.Println("Failed to load rules for current command:", err)
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "specify path to the config file")
	rootCmd.PersistentFlags().StringVar(&rulesDir, "rules-dir", "", "specify path to the rules directory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
