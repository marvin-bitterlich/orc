package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/version"
)

// CheckResult represents the outcome of a single check
type CheckResult struct {
	Name    string
	Status  string // "✓", "⚠", "✗"
	Details string // Only shown if Status != "✓"
}

// DoctorCmd returns the doctor command for environment validation
func DoctorCmd() *cobra.Command {
	var quiet bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate ORC environment and glue deployment",
		Long: `Comprehensive environment health check for ORC.

Validates:
- Directory structure (~/.orc/, ~/wb/)
- ORC repo freshness (commits behind origin/master)
- Glue deployment (skills, hooks, tmux scripts)
- Hook configuration in Claude Code settings
- Binary installation and PATH

Examples:
  orc doctor              # Run full health check
  orc doctor --quiet      # Exit code only (0=healthy, 1=issues)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			results := []CheckResult{}
			hasErrors := false

			// Run all checks
			results = append(results, checkDirectories())
			results = append(results, checkRepoFreshness())

			// Glue checks
			skillResult, hookResult, tmuxResult := checkGlueDeployment()
			results = append(results, skillResult)
			results = append(results, hookResult)
			results = append(results, tmuxResult)

			results = append(results, checkHookConfig())
			results = append(results, checkBinary())

			// Check for errors
			for _, r := range results {
				if r.Status == "✗" {
					hasErrors = true
					break
				}
			}

			if !quiet {
				// Print compact table
				fmt.Println()
				fmt.Println("Check              Status")
				fmt.Println("─────────────────────────")
				for _, r := range results {
					fmt.Printf("%-18s %s\n", r.Name, r.Status)
				}
				fmt.Println()

				// Print details for non-passing checks
				hasDetails := false
				for _, r := range results {
					if r.Status != "✓" && r.Details != "" {
						if !hasDetails {
							fmt.Println("Details:")
							hasDetails = true
						}
						fmt.Printf("\n%s:\n%s\n", r.Name, r.Details)
					}
				}

				if hasErrors {
					fmt.Println("\n⚠ Issues found. Run 'make deploy-glue' to sync glue.")
				} else {
					fmt.Println("All checks passed.")
				}
			}

			if hasErrors {
				return fmt.Errorf("environment validation failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode - exit code only")

	return cmd
}

// checkDirectories validates required directory structure
func checkDirectories() CheckResult {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return CheckResult{Name: "Directories", Status: "✗", Details: "  Cannot get home directory"}
	}

	missing := []string{}

	// Check ~/.orc/
	orcDir := filepath.Join(homeDir, ".orc")
	if _, err := os.Stat(orcDir); os.IsNotExist(err) {
		missing = append(missing, "~/.orc/")
	}

	// Check ~/.orc/ws/
	wsDir := filepath.Join(homeDir, ".orc", "ws")
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		missing = append(missing, "~/.orc/ws/")
	}

	// Check ~/.orc/orc.db
	dbPath := filepath.Join(homeDir, ".orc", "orc.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		missing = append(missing, "~/.orc/orc.db")
	}

	// Check ~/wb/
	wbDir := filepath.Join(homeDir, "wb")
	if _, err := os.Stat(wbDir); os.IsNotExist(err) {
		missing = append(missing, "~/wb/")
	}

	if len(missing) > 0 {
		return CheckResult{
			Name:    "Directories",
			Status:  "✗",
			Details: "  Missing: " + strings.Join(missing, ", "),
		}
	}

	return CheckResult{Name: "Directories", Status: "✓"}
}

// checkRepoFreshness checks if ~/src/orc is behind origin/master
func checkRepoFreshness() CheckResult {
	homeDir, _ := os.UserHomeDir()
	repoPath := filepath.Join(homeDir, "src", "orc")

	// Check if repo exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return CheckResult{
			Name:    "ORC Repo",
			Status:  "⚠",
			Details: "  ~/src/orc not found",
		}
	}

	// Fetch from remote (graceful on failure)
	fetchCmd := exec.Command("git", "-C", repoPath, "fetch", "--quiet")
	fetchCmd.Run() // Ignore errors - network may be unavailable

	// Check commits behind
	revListCmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", "HEAD..origin/master")
	output, err := revListCmd.Output()
	if err != nil {
		return CheckResult{
			Name:    "ORC Repo",
			Status:  "⚠",
			Details: "  Could not check commits (fetch failed?)",
		}
	}

	behind := strings.TrimSpace(string(output))
	if behind != "0" {
		return CheckResult{
			Name:    "ORC Repo",
			Status:  "⚠",
			Details: fmt.Sprintf("  %s commits behind origin/master\n  Run: cd ~/src/orc && git pull", behind),
		}
	}

	return CheckResult{Name: "ORC Repo", Status: "✓"}
}

// checkGlueDeployment compares glue source against deployed locations
func checkGlueDeployment() (skills, hooks, tmux CheckResult) {
	homeDir, _ := os.UserHomeDir()
	glueDir := filepath.Join(homeDir, "src", "orc", "glue")

	// Skills: glue/skills/ -> ~/.claude/skills/
	srcSkills := filepath.Join(glueDir, "skills")
	dstSkills := filepath.Join(homeDir, ".claude", "skills")
	skillsMissing, skillsStale := compareDirs(srcSkills, dstSkills)

	if len(skillsMissing) > 0 || len(skillsStale) > 0 {
		details := ""
		if len(skillsMissing) > 0 {
			details += "  Missing: " + strings.Join(skillsMissing, ", ") + "\n"
		}
		if len(skillsStale) > 0 {
			details += "  Stale: " + strings.Join(skillsStale, ", ")
		}
		skills = CheckResult{Name: "Glue Skills", Status: "✗", Details: strings.TrimSpace(details)}
	} else {
		skills = CheckResult{Name: "Glue Skills", Status: "✓"}
	}

	// Hooks: glue/hooks/ -> ~/.claude/hooks/
	srcHooks := filepath.Join(glueDir, "hooks")
	dstHooks := filepath.Join(homeDir, ".claude", "hooks")
	hooksMissing, hooksStale := compareFiles(srcHooks, dstHooks)

	if len(hooksMissing) > 0 || len(hooksStale) > 0 {
		details := ""
		if len(hooksMissing) > 0 {
			details += "  Missing: " + strings.Join(hooksMissing, ", ") + "\n"
		}
		if len(hooksStale) > 0 {
			details += "  Stale: " + strings.Join(hooksStale, ", ")
		}
		hooks = CheckResult{Name: "Glue Hooks", Status: "✗", Details: strings.TrimSpace(details)}
	} else {
		hooks = CheckResult{Name: "Glue Hooks", Status: "✓"}
	}

	// TMux: glue/tmux/ -> ~/.orc/tmux/
	srcTmux := filepath.Join(glueDir, "tmux")
	dstTmux := filepath.Join(homeDir, ".orc", "tmux")
	tmuxMissing, tmuxStale := compareFiles(srcTmux, dstTmux)

	if len(tmuxMissing) > 0 || len(tmuxStale) > 0 {
		details := ""
		if len(tmuxMissing) > 0 {
			details += "  Missing: " + strings.Join(tmuxMissing, ", ") + "\n"
		}
		if len(tmuxStale) > 0 {
			details += "  Stale: " + strings.Join(tmuxStale, ", ")
		}
		tmux = CheckResult{Name: "Glue TMux", Status: "✗", Details: strings.TrimSpace(details)}
	} else {
		tmux = CheckResult{Name: "Glue TMux", Status: "✓"}
	}

	return skills, hooks, tmux
}

// compareDirs compares directories recursively (for skills)
// Returns lists of missing and stale items
func compareDirs(srcDir, dstDir string) (missing, stale []string) {
	// Check if source exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil, nil // No source, nothing to compare
	}

	// List source directories
	srcEntries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, nil
	}

	for _, entry := range srcEntries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		srcPath := filepath.Join(srcDir, name)
		dstPath := filepath.Join(dstDir, name)

		// Check if destination exists
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			missing = append(missing, name)
			continue
		}

		// Compare contents recursively
		if !dirsEqual(srcPath, dstPath) {
			stale = append(stale, name)
		}
	}

	return missing, stale
}

// dirsEqual compares two directories recursively
func dirsEqual(dir1, dir2 string) bool {
	var files1, files2 []string

	// Collect files from dir1
	_ = filepath.WalkDir(dir1, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir1, path)
		files1 = append(files1, rel)
		return nil
	})

	// Collect files from dir2
	_ = filepath.WalkDir(dir2, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir2, path)
		files2 = append(files2, rel)
		return nil
	})

	// Quick check: same number of files?
	if len(files1) != len(files2) {
		return false
	}

	// Compare each file
	for _, rel := range files1 {
		content1, err1 := os.ReadFile(filepath.Join(dir1, rel))
		content2, err2 := os.ReadFile(filepath.Join(dir2, rel))
		if err1 != nil || err2 != nil || !bytes.Equal(content1, content2) {
			return false
		}
	}

	return true
}

// compareFiles compares files in two directories (for hooks/tmux)
// Returns lists of missing and stale files
func compareFiles(srcDir, dstDir string) (missing, stale []string) {
	// Check if source exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil, nil // No source, nothing to compare
	}

	srcEntries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, nil
	}

	for _, entry := range srcEntries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		srcPath := filepath.Join(srcDir, name)
		dstPath := filepath.Join(dstDir, name)

		// Check if destination exists
		dstInfo, err := os.Stat(dstPath)
		if os.IsNotExist(err) {
			missing = append(missing, name)
			continue
		}
		if err != nil {
			continue
		}

		// Compare contents
		srcContent, err1 := os.ReadFile(srcPath)
		dstContent, err2 := os.ReadFile(dstPath)
		if err1 != nil || err2 != nil {
			continue
		}

		if !bytes.Equal(srcContent, dstContent) {
			stale = append(stale, name)
		}

		_ = dstInfo // Suppress unused warning
	}

	return missing, stale
}

// checkHookConfig verifies glue hooks are configured in settings.json
func checkHookConfig() CheckResult {
	homeDir, _ := os.UserHomeDir()
	glueHooksPath := filepath.Join(homeDir, "src", "orc", "glue", "hooks.json")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Read glue hooks.json
	glueData, err := os.ReadFile(glueHooksPath)
	if err != nil {
		return CheckResult{
			Name:    "Hook Config",
			Status:  "⚠",
			Details: "  Cannot read ~/src/orc/glue/hooks.json",
		}
	}

	var glueHooks map[string]any
	if err := json.Unmarshal(glueData, &glueHooks); err != nil {
		return CheckResult{
			Name:    "Hook Config",
			Status:  "✗",
			Details: "  Invalid JSON in glue/hooks.json",
		}
	}

	// Read settings.json
	settingsData, err := os.ReadFile(settingsPath)
	if err != nil {
		return CheckResult{
			Name:    "Hook Config",
			Status:  "✗",
			Details: "  Cannot read ~/.claude/settings.json",
		}
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		return CheckResult{
			Name:    "Hook Config",
			Status:  "✗",
			Details: "  Invalid JSON in settings.json",
		}
	}

	// Get hooks from settings
	settingsHooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		// No hooks configured at all
		if len(glueHooks) > 0 {
			return CheckResult{
				Name:    "Hook Config",
				Status:  "✗",
				Details: "  No hooks configured in settings.json\n  Run: make deploy-glue",
			}
		}
		return CheckResult{Name: "Hook Config", Status: "✓"}
	}

	// Deep compare each hook type configuration
	var issues []string
	for hookType, glueConfig := range glueHooks {
		settingsConfig, exists := settingsHooks[hookType]
		if !exists {
			issues = append(issues, hookType+": missing")
			continue
		}

		// Serialize both configs for comparison
		glueJSON, _ := json.Marshal(glueConfig)
		settingsJSON, _ := json.Marshal(settingsConfig)

		if string(glueJSON) != string(settingsJSON) {
			issues = append(issues, hookType+": configuration mismatch")
		}
	}

	if len(issues) > 0 {
		return CheckResult{
			Name:    "Hook Config",
			Status:  "✗",
			Details: "  " + strings.Join(issues, "\n  ") + "\n  Run: make deploy-glue",
		}
	}

	return CheckResult{Name: "Hook Config", Status: "✓"}
}

// checkBinary validates orc binary installation
func checkBinary() CheckResult {
	// Check if orc is in PATH
	orcPath, err := exec.LookPath("orc")
	if err != nil {
		return CheckResult{
			Name:    "Binary",
			Status:  "✗",
			Details: "  'orc' not found in PATH\n  Run: make install",
		}
	}

	// If in ORC repo, check local binary freshness
	if isInOrcRepo() {
		localBinary := "./orc"
		if _, err := os.Stat(localBinary); err == nil {
			// Local binary exists, check freshness
			cmd := exec.Command(localBinary, "--version")
			output, err := cmd.Output()
			if err == nil {
				localVersion := strings.TrimSpace(string(output))

				// Get current git commit
				gitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
				gitOutput, err := gitCmd.Output()
				if err == nil {
					currentCommit := strings.TrimSpace(string(gitOutput))
					if !strings.Contains(localVersion, currentCommit) {
						return CheckResult{
							Name:    "Binary",
							Status:  "⚠",
							Details: fmt.Sprintf("  Global: %s\n  Local ./orc is stale (built from different commit)\n  Run: make dev", orcPath),
						}
					}
				}
			}
		}
	}

	return CheckResult{Name: "Binary", Status: "✓", Details: fmt.Sprintf("  %s (%s)", orcPath, version.String())}
}

// isInOrcRepo checks if we're in the ORC repository
func isInOrcRepo() bool {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "module github.com/example/orc")
}
