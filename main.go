package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ecs-manager-ng/common"
	"ecs-manager-ng/ecs"
	p "ecs-manager-ng/prompt"

	"github.com/hako/durafmt"
	"github.com/manifoldco/promptui"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	binName string
	cfgFile string
	cfgDir  string
	version string
	pid     int
)

var (
	result        string
	cluster       string
	instancesInfo []ecs.Instance
	err           error
)

func init() {

	// Use config from ~/.aws
	os.Setenv("AWS_SDK_LOAD_CONFIG", "true")

	// Get process id
	pid = os.Getpid()
	log.Printf("Running %s v%s (PID: %d)", binName, version, pid)

	// Get user's home dir
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf(p.Error("\U00002717 Unable to determine current user's home dir: %s\n\n"), err.Error())
		os.Exit(1)
	}

	// Configuration dir
	cfgDir = filepath.Join(home, ".config/ecs-manager-ng")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(cfgDir)
	}

	// Try to read config
	if err := viper.ReadInConfig(); err != nil {
		log.Printf(p.Error("\U00002717 Unable to read configuration file: %s\n\n"), err.Error())
		os.Exit(1)
	}

	// Usage
	flag.Usage = func() {
		fmt.Printf("Usage: \n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Print config file
	log.Printf("Config file: %s\n\n", viper.ConfigFileUsed())

}

func main() {

	// Main menu
MainMenu:
	prompt := promptui.Select{
		Label: "[ Select action ]",
		Items: []string{
			"Clusters",
			"Quit",
		},
		Size: 30,
	}

	_, result, err = prompt.Run()

	if err != nil {
		fmt.Printf(p.Error("\U00002717 Main menu failed!\n"))
	}

	// Clusters menu
ClustersMenu:
	if result == "Clusters" {

		// Get clusters list
		clusters, err := ecs.GetClusters()

		if err != nil {
			fmt.Printf(p.Error("\U00002717 Couldn't get list of ECS clusters: %v\n"), err)
		}

		// Select cluster
		prompt := promptui.Select{
			Label: "[ Select cluster ]",
			Items: clusters,
			Size:  30,
		}

		_, cluster, err = prompt.Run()

		if err != nil {
			fmt.Printf(p.Error("\U00002717 Cluster selection failed!\n"))
		}

		// Select cluster action
		prompt = promptui.Select{
			Label: "[ Select action ]",
			Items: []string{
				"Instances",
				"Export instances list to file",
				"Update ECS Agent on all instances in cluster",
				"Drain and terminate instances, one by one",
				"Go to clusters menu",
				"Go to main menu",
				"Quit",
			},
			Size: 30,
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf(p.Error("\U00002717 Cluster menu failed!\n"))
		}

		// InstancesMenu:
		if result == "Instances" {
			// Get cluster instances
			instances, err := ecs.GetClusterInstances(cluster)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), cluster, err)
			}

			if len(instances) > 0 {
				instancesInfo, err = ecs.GetClusterInstancesInfo(cluster, instances)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), cluster, err)
				}

				templates := &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "\U00002771 {{ .Name | blue }} [ ami:{{ .AMI | cyan }} | r:{{ .RunningTasksCount | cyan }} | p:{{ .PendingTasksCount | cyan }} | agent:{{ .AgentVersion | cyan }} ] \U00002770",
					Inactive: "  {{ .Name | blue }} [ ami:{{ .AMI | cyan }} | r:{{ .RunningTasksCount | cyan }} | p:{{ .PendingTasksCount | cyan }} | agent:{{ .AgentVersion | cyan }} ]",
					Selected: "\U00002714 {{ .Name | blue }} [ ami:{{ .AMI | cyan }} | r:{{ .RunningTasksCount | cyan }} | p:{{ .PendingTasksCount | cyan }} | agent:{{ .AgentVersion | cyan }} ]",
					Details: `
			--------------[ Instance details ]---------------
			{{ "ARN:" | faint }}              {{ .ARN }}
			{{ "Status:" | faint }}           {{ .Status }}
			{{ "EC2 Instance ID:" | faint }}  {{ .Ec2InstanceID }}
			{{ "AMI:" | faint }}              {{ .AMI }}
			{{ "Agent Version:" | faint }}    {{ .AgentVersion }}
			{{ "Docker Version:" | faint }}   {{ .DockerVersion }}
			{{ "Running Tasks:" | faint }}    {{ .RunningTasksCount }}
			{{ "Pending Tasks:" | faint }}    {{ .PendingTasksCount }}`,
				}

				searcher := func(input string, index int) bool {
					inst := instancesInfo[index]
					name := strings.Replace(strings.ToLower(inst.Name), " ", "", -1)
					input = strings.Replace(strings.ToLower(input), " ", "", -1)

					return strings.Contains(name, input)
				}

				prompt = promptui.Select{
					Label:     "Select instance",
					Items:     instancesInfo,
					Templates: templates,
					Size:      10,
					Searcher:  searcher,
				}

				_, result, err = prompt.Run()

				if err != nil {
					fmt.Printf(p.Error("\U00002717 Prompt failed %v\n"), err)
				}

				prompt = promptui.Select{
					Label: "[ Select action ]",
					Items: []string{
						"Update ECS Agent",
						"Activate instance",
						"Drain instance",
						"Terminate instance",
						"Drain and terminate instance",
						"Go to clusters menu",
						"Go to main menu",
						"Quit",
					},
				}

				_, result, err = prompt.Run()

				if err != nil {
					fmt.Printf(p.Error("\U00002717 Prompt failed %v\n"), err)
					os.Exit(0)
				}

				if result == "Quit" {
					os.Exit(0)
				}

			} else {
				fmt.Println(p.Info("\U00002717 No instances in cluster."))
			}

		}

		// Export instances list to file
		if result == "Export instances list to file" {
			startTime := time.Now()

			// Get cluster instances
			instances, err := ecs.GetClusterInstances(cluster)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n", cluster, err))
			}

			if len(instances) > 0 {

				instancesInfo, err = ecs.GetClusterInstancesInfo(cluster, instances)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), cluster, err)
				}

				// File name
				filename := filepath.Join(cfgDir, fmt.Sprintf("%s-instances.list", strings.Split(cluster, "/")[1]))

				// Create file and open for writing
				f, err := os.Create(filename)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't create file %s: %v\n"), filename, err)
				}

				// Iterate through instances and write to file
				for _, inst := range instancesInfo {
					line := fmt.Sprintf("%s (EC2:%s, AMI:%s)\n", inst.Name, inst.Ec2InstanceID, inst.AMI)
					_, err := f.WriteString(line)
					if err != nil {
						fmt.Printf(p.Error("\U00002717 Couldn't write to file %s: %v\n"), filename, err)
						f.Close()
						goto ClustersMenu
					}
				}
				f.Close()

				fmt.Printf(p.Info("\U00002714 List exported to %s\n"), filename)

			} else {
				fmt.Println(p.Info("\U00002717 No instances in cluster, nothing to export."))
			}

			// Calculate elapsed time and print it
			elapsedTime := time.Since(startTime)
			duration := durafmt.ParseShort(elapsedTime)
			fmt.Printf("\n_____________________________________________\n\n")
			fmt.Printf("   %s %s\n", p.Grey("Duration:"), duration.String())
			fmt.Printf("_____________________________________________\n\n")
			goto ClustersMenu
		}

		// Update ECS Agent on all instances in cluster
		if result == "Update ECS Agent on all instances in cluster" {
			startTime := time.Now()

			// Get cluster instances
			instances, err := ecs.GetClusterInstances(cluster)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), cluster, err)
			}

			if len(instances) > 0 {
				instancesInfo, err = ecs.GetClusterInstancesInfo(cluster, instances)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), cluster, err)
				}

				// Iterate through instance list and update container agent
				for i, inst := range instancesInfo {
					fmt.Printf(p.Info("\U0001F5A5  [%d/%d] %s (%s): "), i+1, len(instancesInfo), inst.Name, inst.Ec2InstanceID)
					r, e := ecs.UpdateContainerAgent(cluster, inst.Name)
					if e != nil {
						fmt.Printf(p.Error("\U00002717 Couldn't update container agent: %v"), err)
					}
					fmt.Println(p.Yellow(r))

					// Let's wait a few seconds before proceeding to next instance
					if r == "PENDING" && i < len(instancesInfo)-1 {
						time.Sleep(10 * time.Second)
					}
				}
			} else {
				fmt.Println(p.Info("\U00002717 No instances in cluster, nothing to do."))
			}

			// Calculate elapsed time and print it
			elapsedTime := time.Since(startTime)
			duration := durafmt.ParseShort(elapsedTime)
			fmt.Printf("\n_____________________________________________\n\n")
			fmt.Printf("   %s %s\n", p.Grey("Duration:"), duration.String())
			fmt.Printf("_____________________________________________\n\n")

			goto ClustersMenu
		}

		// Drain and terminate instances, one by one
		if result == "Drain and terminate instances, one by one" {
			prompt := promptui.Prompt{
				Label:     "Are you sure you want to do this",
				IsConfirm: true,
			}

			result, err := prompt.Run()

			if err != nil || result != "y" {
				os.Exit(0)
			}

			startTime := time.Now()

			// Get cluster instances
			instances, err := ecs.GetClusterInstances(cluster)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), cluster, err)
			}

			// Get list of excluded instances
			excludeFilename := filepath.Join(cfgDir, fmt.Sprintf("%s-instances.exclude", strings.Split(cluster, "/")[1]))
			excludedInstances, err := common.ReadFileLines(excludeFilename)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of excluded instances from %s: %v\n"), excludeFilename, err)
			}

			// Get cluster info
			r, err := ecs.GetClusterInfo(cluster)
			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get cluster info: %v\n"), cluster, err)
			}
			registeredInstancesCount := r.RegisteredInstancesCount

			// Iterate over instances
			if len(instances) > 0 {
				instancesInfo, err = ecs.GetClusterInstancesInfo(cluster, instances)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), cluster, err)
				}

				// Iterate through instance list and update container agent
				for i, inst := range instancesInfo {
					fmt.Printf(p.Info("\U0001F5A5  [%d/%d] Instance %s: "), i+1, len(instancesInfo), inst.Name)

					// Check if instance is excluded
					if common.ElementInSlice(inst.Name, excludedInstances) {
						fmt.Println(p.Green("EXCLUDED"))
						continue
					}

					// Drain instance
					r, err := ecs.DrainContainerInstance(cluster, inst.Name)
					if err != nil {
						fmt.Printf(p.Error("\n   \U00002717 Couldn't drain container instance: %v\n"), err)
					}
					fmt.Println(p.Yellow(r))

					// Get instance info
					r1, err := ecs.GetClusterInstancesInfo(cluster, []string{inst.Name})
					if err != nil {
						fmt.Printf(p.Error("\n   \U00002717 Couldn't get instance info: %v\n"), err)
					}
					inst := r1[0]

					loop := true
					for loop {
						sleepTime := 10 * time.Second

						// Get instance task list
						tasks, err := ecs.GetInstanceTasks(cluster, inst.Name)

						if err != nil {
							fmt.Printf(p.Error("\n   \U00002717 Couldn't get list of tasks: %v\n"), err)
						}

						// Number of tasks
						runningTasksCount := len(tasks)

						fmt.Printf("\r   \U0000276F %s %s (need %s)  ", p.Grey("Running tasks:"), p.Green(runningTasksCount), p.Yellow("0"))

						// If task count reaced 0, stop the loop
						if runningTasksCount == 0 {
							loop = false
						}

						if loop {
							time.Sleep(sleepTime)
						}
					}
					fmt.Println()

					fmt.Printf(p.Info("   \U0000276F Terminate instance %s (%s): "), inst.Name, inst.Ec2InstanceID)

					// Terminate instance
					r, err = ecs.TerminateContainerInstance(inst.Ec2InstanceID)
					if err != nil {
						fmt.Printf(p.Error("\n   \U00002717 Couldn't terminate instance %s (%s): %v\n"), inst.Name, inst.Ec2InstanceID, err)
					}
					fmt.Println(p.Yellow(r))
					fmt.Printf("   \U0000276F %s\n", p.Grey("Waiting for instance to shut down."))

					loop = true
					for loop {
						sleepTime := 10 * time.Second

						// Get cluster info
						r, err := ecs.GetClusterInfo(cluster)

						if err != nil {
							fmt.Printf(p.Error("\n   \U00002717 Couldn't get cluster info: %v\n"), err)
						}

						fmt.Printf("\r   \U0000276F %s %s (need %s)  ", p.Grey("Registered instances count:"), p.Green(r.RegisteredInstancesCount), p.Yellow(registeredInstancesCount-1))

						// If registered instances count is decresed by one (one instance terminated), stop the loop
						if r.RegisteredInstancesCount == registeredInstancesCount-1 {
							loop = false
						}

						if loop {
							time.Sleep(sleepTime)
						}
					}
					fmt.Printf("\n   \U0000276F %s\n", p.Grey("Instance terminated. Waiting for a new one."))

					loop = true
					for loop {
						sleepTime := 10 * time.Second

						// Get cluster info
						r, err := ecs.GetClusterInfo(cluster)

						if err != nil {
							fmt.Printf(p.Error("\n   \U00002717 Couldn't get cluster info: %v\n"), err)
						}

						fmt.Printf("\r   \U0000276F %s %s (need %s)  ", p.Grey("Registered instances count:"), p.Green(r.RegisteredInstancesCount), p.Yellow(registeredInstancesCount))

						// If registered instances count is back to initial value (all instances in cluster), stop the loop
						if r.RegisteredInstancesCount >= registeredInstancesCount {
							loop = false
						}

						if loop {
							time.Sleep(sleepTime)
						}
					}
					fmt.Println()
				}

				// Calculate elapsed time and print it
				elapsedTime := time.Since(startTime)
				duration := durafmt.ParseShort(elapsedTime)
				fmt.Printf("\n_____________________________________________\n\n")
				fmt.Printf("   %s %s\n", p.Grey("Duration:"), duration.String())
				fmt.Printf("_____________________________________________\n\n")

			} else {
				fmt.Println(p.Info("\U00002717 No instances in cluster, nothing to do."))
			}

			goto ClustersMenu
		}

		// Jump to clusters menu
		if result == "Go to clusters menu" {
			goto ClustersMenu
		}

		// Jump to main menu
		if result == "Go to main menu" {
			goto MainMenu
		}

		// Quit
		if result == "Quit" {
			os.Exit(0)
		}
	}

	if result == "Go to main menu" {
		goto MainMenu
	}

	if result == "Quit" {
		os.Exit(0)
	}

}
