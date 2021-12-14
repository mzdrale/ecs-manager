package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/mzdrale/ecs-manager/aws"
	"gitlab.com/mzdrale/ecs-manager/common"

	p "gitlab.com/mzdrale/ecs-manager/prompt"

	"github.com/briandowns/spinner"
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
	result           string
	cluster          string
	ecsInstancesInfo []aws.EcsInstance
	err              error
)

// Config variables
var (
	aPrintVersion              bool
	numberOfZeroTasksInstances int
)

func init() {

	// Use config from ~/.aws
	os.Setenv("AWS_SDK_LOAD_CONFIG", "true")

	// Get user's home dir
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf(p.Error("\U00002717 Unable to determine current user's home dir: %s\n\n"), err.Error())
		os.Exit(1)
	}

	// Configuration dir
	cfgDir = filepath.Join(home, ".config/ecs-manager")

	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgDir)

	// Try to read config
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf(p.Error("\U00002717 Unable to read configuration file: %s\n\n"), err.Error())
		os.Exit(1)
	}

	// Usage
	flag.Usage = func() {
		fmt.Printf("Usage: \n")
		flag.PrintDefaults()
	}

	// Get arguments
	flag.BoolVarP(&aPrintVersion, "version", "V", false, "Print version")

	flag.Parse()

}

func main() {

	if aPrintVersion {
		fmt.Printf("\n%v %v\n\n", binName, version)
		fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
		fmt.Printf("URL: https://gitlab.com/mzdrale/ecs-manager\n\n")
		os.Exit(0)
	}

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
		clusters, err := aws.GetEcsClusters()

		if err != nil {
			fmt.Printf(p.Error("\U00002717 Couldn't get list of ECS clusters: %v\n"), err)
		}

		// If no clusters found, return to main menu
		if len(clusters) == 0 {
			fmt.Println(p.Info("\U00002717 No clusters found."))
			goto MainMenu
		}

		clustersInfo, err := aws.GetEcsClustersInfo(clusters)
		if err != nil {
			fmt.Printf(p.Error("\U00002717 Couldn't get list of ECS clusters: %v\n"), err)
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U00002771 {{ .Name | blue }} [ instances:{{ .RegisteredInstancesCount | cyan }} | running:{{ .RunningTasksCount | cyan }} | pending:{{ .PendingTasksCount | cyan }} ] \U00002770",
			Inactive: "  {{ .Name | blue }} [ instances:{{ .RegisteredInstancesCount | cyan }} | running:{{ .RunningTasksCount | cyan }} | pending:{{ .PendingTasksCount | cyan }} ]",
			Selected: "\U00002714 {{ .Name | blue }} [ instances:{{ .RegisteredInstancesCount | cyan }} | running:{{ .RunningTasksCount | cyan }} | pending:{{ .PendingTasksCount | cyan }} ]",
			Details: `
		--------------[ Cluster details ]---------------
		{{ "ARN:" | faint }}              {{ .ARN }}
		{{ "Name:" | faint }}             {{ .Name }}
		{{ "Status:" | faint }}           {{ .Status }}
		{{ "Instances Count:" | faint }}  {{ .RegisteredInstancesCount }}
		{{ "Running Tasks:" | faint }}    {{ .RunningTasksCount }}
		{{ "Pending Tasks:" | faint }}    {{ .PendingTasksCount }}`,
		}

		searcher := func(input string, index int) bool {
			clust := clustersInfo[index]
			name := strings.Replace(strings.ToLower(clust.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt = promptui.Select{
			Label:     "Select cluster",
			Items:     clustersInfo,
			Templates: templates,
			Size:      10,
			Searcher:  searcher,
		}

		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf(p.Error("\U00002717 Prompt failed %v\n"), err)
		}

		clust := clustersInfo[i]

		testCluster := viper.GetBool(fmt.Sprintf("ecs.%s.test_cluster", clust.ARN))
		waitForTaskCluster := viper.GetBool(fmt.Sprintf("ecs.%s.wait_for_task", clust.ARN))
		numberOfZeroTasksInstances = 0

		if testCluster {
			fmt.Printf(p.Red("\n==============================================================\n"))
			fmt.Printf(p.Red("                          TEST CLUSTER \n"))
			fmt.Printf(p.Red("______________________________________________________________\n\n"))
			fmt.Printf(p.Red(" This cluster is marked as test cluster (check config file). \n"))
			fmt.Printf(p.Red(" It means if you chose to drain instances in this cluster, \n"))
			fmt.Printf(p.Red(" this tool would not wait for drain to finish, but force stop \n"))
			fmt.Printf(p.Red(" tasks one by one.\n"))
			fmt.Printf(p.Red("______________________________________________________________\n\n"))
		}

		if waitForTaskCluster {
			numberOfZeroTasksInstances = viper.GetInt(fmt.Sprintf("ecs.%s.number_of_zero_tasks_instances", clust.ARN))
			fmt.Printf(p.Magenta("\n==============================================================\n"))
			fmt.Printf(p.Magenta("                     WAIT FOR TASK CLUSTER \n"))
			fmt.Printf(p.Magenta("______________________________________________________________\n\n"))
			fmt.Printf(p.Magenta(" This cluster is configured to wait for task (check config file). \n"))
			fmt.Printf(p.Magenta(" It means if you chose to drain and terminate instances\n"))
			fmt.Printf(p.Magenta(" in this cluster, this tool would wait for a new instance\n"))
			fmt.Printf(p.Magenta(" to come up and start at least one task before proceeding\n"))
			fmt.Printf(p.Magenta(" to the next one.\n"))
			fmt.Printf(p.Magenta(" Allowed number of instances with 0 tasks running: ", p.Yellow(numberOfZeroTasksInstances)))
			fmt.Printf(p.Magenta("\n______________________________________________________________\n\n"))
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

	InstancesMenu:
		if result == "Instances" {
			// Get cluster instances
			instances, err := aws.GetEcsClusterInstances(clust.ARN)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), clust.Name, err)
			}

			if len(instances) > 0 {
				ecsInstancesInfo, err = aws.GetEcsClusterInstancesInfo(clust.ARN, instances)

				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), clust.Name, err)
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
					inst := ecsInstancesInfo[index]
					name := strings.Replace(strings.ToLower(inst.Name), " ", "", -1)
					input = strings.Replace(strings.ToLower(input), " ", "", -1)

					return strings.Contains(name, input)
				}

				prompt = promptui.Select{
					Label:     "Select instance",
					Items:     ecsInstancesInfo,
					Templates: templates,
					Size:      10,
					Searcher:  searcher,
				}

				i, result, err := prompt.Run()

				if err != nil {
					fmt.Printf(p.Error("\U00002717 Prompt failed %v\n"), err)
				}

				inst := ecsInstancesInfo[i]

				prompt = promptui.Select{
					Label: "[ Select action ]",
					Items: []string{
						"Update ECS Agent",
						"Activate instance",
						"Drain instance",
						"Terminate instance",
						"Drain and terminate instance",
						"Go to instances menu",
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

				// Update ECS Agent
				if result == "Update ECS Agent" {
					startTime := time.Now()

					fmt.Printf(p.Info("\U0001F5A5  Update ECS Agent on %s (%s): "), inst.Name, inst.Ec2InstanceID)
					r, err := aws.UpdateEcsContainerAgent(clust.ARN, inst.Name)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't update container agent: %v"), err)
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Calculate elapsed time and print it
					elapsedTime := time.Since(startTime)
					fmt.Printf("\n_____________________________________________\n\n")
					fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
					fmt.Printf("_____________________________________________\n\n")
					goto InstancesMenu

				}

				// Activate instance
				if result == "Activate instance" {
					startTime := time.Now()

					fmt.Printf(p.Info("\U0001F5A5  Activate instance %s (%s): "), inst.Name, inst.Ec2InstanceID)
					r, err := aws.ActivateEcsContainerInstance(clust.ARN, inst.ARN)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't activate instance: %v"), err)
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Calculate elapsed time and print it
					elapsedTime := time.Since(startTime)
					fmt.Printf("\n_____________________________________________\n\n")
					fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
					fmt.Printf("_____________________________________________\n\n")
					goto InstancesMenu
				}

				// Drain instance
				if result == "Drain instance" {
					startTime := time.Now()

					fmt.Printf(p.Info("\U0001F5A5  Drain instance %s (%s): "), inst.Name, inst.Ec2InstanceID)
					r, err := aws.DrainEcsContainerInstance(clust.ARN, inst.ARN)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't drain instance: %v"), err)
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Calculate elapsed time and print it
					elapsedTime := time.Since(startTime)
					fmt.Printf("\n_____________________________________________\n\n")
					fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
					fmt.Printf("_____________________________________________\n\n")
					goto InstancesMenu

				}

				// Terminate instance
				if result == "Terminate instance" {
					prompt := promptui.Prompt{
						Label:     "Are you sure you want to do this",
						IsConfirm: true,
					}

					result, err := prompt.Run()

					if err != nil || result != "y" {
						goto InstancesMenu
					}

					startTime := time.Now()

					fmt.Printf(p.Info("\U0001F5A5  Terminate instance %s (%s): "), inst.Name, inst.Ec2InstanceID)
					r, err := aws.TerminateEc2Instance(inst.Ec2InstanceID)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't terminate instance: %v"), err)
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Calculate elapsed time and print it
					elapsedTime := time.Since(startTime)
					fmt.Printf("\n_____________________________________________\n\n")
					fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
					fmt.Printf("_____________________________________________\n\n")

					// Sleep few seconds before going back to instances list
					time.Sleep(3 * time.Second)
					goto InstancesMenu
				}

				// Drain and terminate instance
				if result == "Drain and terminate instance" {
					prompt := promptui.Prompt{
						Label:     "Are you sure you want to do this",
						IsConfirm: true,
					}

					result, err := prompt.Run()

					if err != nil || result != "y" {
						goto InstancesMenu
					}

					startTime := time.Now()

					// Drain instance
					fmt.Printf(p.Info("\U0001F6B0 Drain instance %s (%s): "), inst.Name, inst.Ec2InstanceID)
					r, err := aws.DrainEcsContainerInstance(clust.ARN, inst.Name)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't drain container instance: %v\n"), err)
						goto InstancesMenu
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Get instance info
					r1, err := aws.GetEcsClusterInstancesInfo(clust.ARN, []string{inst.Name})
					if err != nil {
						fmt.Printf(p.Error("\n   \U00002717 Couldn't get instance info: %v\n"), err)
					}
					inst := r1[0]

					loop := true
					actionFailedCnt := 0
					for loop {
						sleepTime := 10 * time.Second

						// Get instance task list
						tasks, err := aws.GetEcsInstanceTasks(clust.ARN, inst.Name)

						if err != nil {
							fmt.Printf(p.Error("\n   \U00002717 Couldn't get list of tasks: %v\n"), err)
						}

						// Number of tasks
						runningTasksCount := len(tasks)

						// If task count reached 0, stop the loop
						if runningTasksCount == 0 {
							loop = false
						}

						// If it's test cluster, stop tasks, don't wait for drain to finish
						if testCluster && runningTasksCount > 0 {
							r, err = aws.StopEcsTask(clust.ARN, tasks[0])
							fmt.Printf(p.Info("   \U0000276F Stop task %s: "), tasks[0])
							if err != nil {
								fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't stop the task: %v\n"), err)
								actionFailedCnt++
							} else {
								fmt.Println(p.Yellow(r))
							}
						} else {
							fmt.Printf("\r   \U0000276F %s %s (need %s)  ", p.Grey("Running tasks:"), p.Green(runningTasksCount), p.Yellow("0"))
						}

						// If action failed so many times, give up
						if actionFailedCnt > 5 {
							fmt.Print(p.Error("    \U00002937 Failed too many times, giving up!\n\n"))
							goto InstancesMenu
						}

						if loop && r != "FAILED" {
							time.Sleep(sleepTime)
						}
					}
					fmt.Println()

					fmt.Printf(p.Info("   \U0000276F Terminate instance: "))

					// Terminate instance
					r, err = aws.TerminateEc2Instance(inst.Ec2InstanceID)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't terminate instance: %v"), err)
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Calculate elapsed time and print it
					elapsedTime := time.Since(startTime)
					fmt.Printf("\n_____________________________________________\n\n")
					fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
					fmt.Printf("_____________________________________________\n\n")

					// Sleep few seconds before going back to instances list
					time.Sleep(3 * time.Second)
					goto InstancesMenu
				}

				if result == "Go to instances menu" {
					goto InstancesMenu
				}

				if result == "Go to clusters menu" {
					goto ClustersMenu
				}

				if result == "Go to main menu" {
					goto MainMenu
				}

				if result == "Quit" {
					os.Exit(0)
				}

			} else {
				fmt.Println(p.Info("\U00002717 No instances in cluster."))
				goto ClustersMenu
			}

		}

		// Export instances list to file
		if result == "Export instances list to file" {
			startTime := time.Now()

			// Get cluster instances
			instances, err := aws.GetEcsClusterInstances(clust.ARN)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n", clust.Name, err))
			}

			if len(instances) > 0 {

				ecsInstancesInfo, err = aws.GetEcsClusterInstancesInfo(clust.ARN, instances)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), clust.Name, err)
				}

				// File name
				filename := filepath.Join(cfgDir, fmt.Sprintf("%s-instances.list", clust.Name))

				// Create file and open for writing
				f, err := os.Create(filename)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't create file %s: %v\n"), filename, err)
				}

				// Iterate through instances and write to file
				for _, inst := range ecsInstancesInfo {
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
			fmt.Printf("\n_____________________________________________\n\n")
			fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
			fmt.Printf("_____________________________________________\n\n")
			goto ClustersMenu
		}

		// Update ECS Agent on all instances in cluster
		if result == "Update ECS Agent on all instances in cluster" {
			startTime := time.Now()

			// Get cluster instances
			instances, err := aws.GetEcsClusterInstances(clust.ARN)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), clust.Name, err)
			}

			if len(instances) > 0 {
				ecsInstancesInfo, err = aws.GetEcsClusterInstancesInfo(clust.ARN, instances)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), clust.Name, err)
				}

				// Iterate through instance list and update container agent
				for i, inst := range ecsInstancesInfo {
					fmt.Printf(p.Info("\U0001F5A5  Update agent on %s (%s) [%02d/%02d]: "), inst.Name, inst.Ec2InstanceID, i+1, len(ecsInstancesInfo))
					r, err := aws.UpdateEcsContainerAgent(clust.ARN, inst.Name)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n    \U00002937 \U00002717 Couldn't update container agent: %v\n"), err)
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Let's wait a few seconds before proceeding to next instance
					if r == "PENDING" && i < len(ecsInstancesInfo)-1 {
						time.Sleep(10 * time.Second)
					}
				}
			} else {
				fmt.Println(p.Info("\U00002717 No instances in cluster, nothing to do."))
			}

			// Calculate elapsed time and print it
			elapsedTime := time.Since(startTime)
			fmt.Printf("\n_____________________________________________\n\n")
			fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
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
				goto ClustersMenu
			}

			startTime := time.Now()

			// Get cluster instances
			instances, err := aws.GetEcsClusterInstances(clust.ARN)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), clust.Name, err)
			}

			// Get list of excluded instances
			excludeFilename := filepath.Join(cfgDir, fmt.Sprintf("%s-instances.exclude", clust.Name))
			excludedInstances, err := common.ReadExcludedInstancesList(excludeFilename)

			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get list of excluded instances from %s: %v\n"), excludeFilename, err)
			}

			// If there are instances in excluded list, raise a warning
			if len(excludedInstances) > 0 {
				fmt.Printf(p.Warn("\U000026A0 Exclude list is not empty: %s\n"), strings.Join(excludedInstances, ", "))

				prompt := promptui.Prompt{
					Label:     "Do you want to exclude these instances",
					IsConfirm: true,
				}

				result, err := prompt.Run()

				if err != nil || result != "y" {
					excludedInstances = []string{}
				}

			}

			// Get cluster info
			r, err := aws.GetEcsClustersInfo([]string{clust.ARN})
			if err != nil {
				fmt.Printf(p.Error("\U00002717 Couldn't get cluster info: %v\n"), clust.Name, err)
			}

			registeredInstancesCount := r[0].RegisteredInstancesCount

			// Iterate over instances
			if len(instances) > 0 {
				ecsInstancesInfo, err = aws.GetEcsClusterInstancesInfo(clust.ARN, instances)
				if err != nil {
					fmt.Printf(p.Error("\U00002717 Couldn't get list of instances in ECS cluster %s: %v\n"), clust.Name, err)
				}

				s := spinner.New(spinner.CharSets[11], 200*time.Millisecond)

				// Iterate through instance list and drain and terminate instances
				for i, inst := range ecsInstancesInfo {
					fmt.Printf(p.Info("\U0001F5A5  [%02d/%02d] Instance %s (%s)\n"), i+1, len(ecsInstancesInfo), inst.Name, inst.Ec2InstanceID)

					if waitForTaskCluster {
						loop := true

						s.Prefix = fmt.Sprintf("   \U0000276F %s ", p.Grey("Waiting for instances to get in active state and start task(s)"))
						s.Start()

						for loop {
							sleepTime := 10 * time.Second

							if aws.IsEcsClusterReady(clust.ARN, true, numberOfZeroTasksInstances) {
								s.Stop()
								fmt.Printf("   \U0000276F %s \n", p.Grey("Waiting for instances to get in active state and start task(s)"))
								fmt.Printf("   \U0000276F %s \n", p.Grey("All instances are active and running at least one task"))
								loop = false
							}

							if loop {
								time.Sleep(sleepTime)
							}
						}
						s.Stop()
					}

					fmt.Printf(p.Info("   \U0000276F Drain instance: "))

					// Check if instance is excluded
					if common.ElementInSlice(inst.Name, excludedInstances) {
						fmt.Println(p.Green("EXCLUDED"))
						continue
					}

					// Drain instance
					r, err := aws.DrainEcsContainerInstance(clust.ARN, inst.Name)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n      \U00002937 \U00002717 Couldn't drain container instance: %v\n"), err)
						continue
					} else {
						fmt.Println(p.Yellow(r))
					}

					// Get instance info
					r1, err := aws.GetEcsClusterInstancesInfo(clust.ARN, []string{inst.Name})
					if err != nil {
						fmt.Printf(p.Error("\n   \U00002717 Couldn't get instance info: %v\n"), err)
					}
					inst := r1[0]

					loop := true
					for loop {
						sleepTime := 10 * time.Second

						// Get instance task list
						tasks, err := aws.GetEcsInstanceTasks(clust.ARN, inst.Name)

						if err != nil {
							fmt.Printf(p.Error("\n   \U00002717 Couldn't get list of tasks: %v\n"), err)
						}

						// Number of tasks
						runningTasksCount := len(tasks)

						// If it's test cluster, stop tasks, don't wait for drain to finish
						if testCluster && runningTasksCount > 0 {
							r, err = aws.StopEcsTask(clust.ARN, tasks[0])
							fmt.Printf(p.Info("   \U0000276F Stop task %s: "), tasks[0])

							if err != nil {
								fmt.Printf(p.Error("FAILED\n      \U00002937 \U00002717 Couldn't stop the task: %v\n"), err)
							} else {
								fmt.Println(p.Yellow(r))
							}
						} else {
							fmt.Printf("\r   \U0000276F %s %s (need %s)  ", p.Grey("Running tasks:"), p.Green(runningTasksCount), p.Yellow("0"))
						}

						// If task count reached 0, stop the loop
						if runningTasksCount == 0 {
							loop = false
						}

						if loop {
							time.Sleep(sleepTime)
						}
					}
					fmt.Println()

					fmt.Printf(p.Info("   \U0000276F Terminate instance: "))

					// Terminate instance
					r, err = aws.TerminateEc2Instance(inst.Ec2InstanceID)
					if err != nil {
						fmt.Printf(p.Error("FAILED\n      \U00002937 \U00002717 Couldn't terminate instance: %v\n"), err)
					} else {
						fmt.Println(p.Yellow(r))
					}

					s.Prefix = fmt.Sprintf("   \U0000276F %s ", p.Grey("Waiting for instance to shut down"))
					s.Start()

					// Wait for number of registered instances to decrease by 1
					loop = true
					for loop {
						sleepTime := 10 * time.Second

						if aws.IsEc2InstanceTerminated(inst.Ec2InstanceID) {
							s.Stop()
							fmt.Printf("   \U0000276F %s \n", p.Grey("Waiting for instance to shut down"))
							loop = false
						}

						if loop {
							time.Sleep(sleepTime)
						}
					}
					fmt.Printf("   \U0000276F %s\n", p.Grey("Instance terminated, waiting for a new one"))
					s.Stop()

					// Wait for number of registered instances to go back to initial value
					loop = true
					for loop {
						sleepTime := 10 * time.Second

						// Get cluster info
						r, err := aws.GetEcsClustersInfo([]string{clust.ARN})

						if err != nil {
							fmt.Printf(p.Error("\n   \U00002717 Couldn't get cluster info: %v\n"), err)
						}

						fmt.Printf("\r   \U0000276F %s %s (need %s)  ", p.Grey("Registered instances count:"), p.Green(r[0].RegisteredInstancesCount), p.Yellow(registeredInstancesCount))

						// If registered instances count is back to initial value (all instances in cluster), stop the loop
						if r[0].RegisteredInstancesCount >= registeredInstancesCount {
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
				fmt.Printf("\n_____________________________________________\n\n")
				fmt.Printf("   %s %s\n", p.Grey("Duration:"), common.FormatDuration(elapsedTime))
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
