package templates

import (
	"embed"
)

//go:embed prime/*.tmpl
var primeTemplates embed.FS

// GetCoreRules returns the core rules template content
func GetCoreRules() (string, error) {
	content, err := primeTemplates.ReadFile("prime/core-rules.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetGitDiscovery returns the git discovery template content
func GetGitDiscovery() (string, error) {
	content, err := primeTemplates.ReadFile("prime/git-discovery.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetWelcomeGoblin returns the Goblin welcome message template content
func GetWelcomeGoblin() (string, error) {
	content, err := primeTemplates.ReadFile("prime/welcome-goblin.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetWelcomeIMP returns the IMP welcome message template content
func GetWelcomeIMP() (string, error) {
	content, err := primeTemplates.ReadFile("prime/welcome-imp.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetWelcomeWatchdog returns the Watchdog welcome message template content
func GetWelcomeWatchdog() (string, error) {
	content, err := primeTemplates.ReadFile("prime/welcome-watchdog.tmpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
