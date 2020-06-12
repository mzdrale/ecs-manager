package prompt

import (
	"os"

	"github.com/manifoldco/promptui"
)

// PrintMainMenu - print main menu
func PrintMainMenu() (int, string, error) {
	prompt := promptui.Select{
		Label: "[ Select action ]",
		Items: []string{
			"Clusters",
			"Quit",
		},
		Size: 30,
	}

	i, result, err := prompt.Run()
	if err != nil {
		return -1, "", err
	}

	if result == "Quit" {
		os.Exit(0)
	}

	return i, result, err
}

// PrintClusterMenu - print cluster menu
func PrintClusterMenu() (int, string, error) {
	prompt := promptui.Select{
		Label: "[ Select action ]",
		Items: []string{
			"Instances",
			"Export instances list to file",
			"Update ECS Agent on all instances in cluster",
			"Drain and terminate instances, one by one",
			"Go to main menu",
			"Quit",
		},
		Size: 30,
	}

	i, result, err := prompt.Run()

	if err != nil {
		return -1, "", err
	}
	return i, result, err
}
