package aws

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// EcsInstance holds information about ECS instance
type EcsInstance struct {
	ARN               string
	Name              string
	Ec2InstanceID     string
	AMI               string
	Status            string
	AgentVersion      string
	DockerVersion     string
	PendingTasksCount int64
	RunningTasksCount int64
	RegisteredAt      string
	RemainingCPU      int64
	RemainingMemory   int64
}

// EcsCluster holds information about ECS cluster
type EcsCluster struct {
	ARN                      string
	Name                     string
	Status                   string
	Region                   string
	Account                  string
	RegisteredInstancesCount int64
	RunningTasksCount        int64
	PendingTasksCount        int64
	ActiveServicesCount      int64
}

// GetEcsClusters - gets list of ECS clusters
func GetEcsClusters() ([]string, error) {
	clusters := []string{}

	svc := ecs.New(session.New())

	input := &ecs.ListClustersInput{}
	result, err := svc.ListClusters(input)

	if err != nil {
		return clusters, err
	}

	for _, clusterArn := range result.ClusterArns {
		clusters = append(clusters, *clusterArn)
	}

	return clusters, nil
}

// GetEcsClustersInfo - gets ECS clusters info
func GetEcsClustersInfo(arns []string) ([]EcsCluster, error) {
	clusterInfo := EcsCluster{}
	clustersInfo := []EcsCluster{}

	svc := ecs.New(session.New())

	input := &ecs.DescribeClustersInput{
		Clusters: aws.StringSlice(arns),
	}

	result, err := svc.DescribeClusters(input)

	if err != nil {
		return clustersInfo, err
	}

	for _, r := range result.Clusters {
		clusterInfo.ARN = *r.ClusterArn
		clusterInfo.Name = *r.ClusterName
		clusterInfo.Status = *r.Status
		clusterInfo.RegisteredInstancesCount = *r.RegisteredContainerInstancesCount
		clusterInfo.RunningTasksCount = *r.RunningTasksCount
		clusterInfo.PendingTasksCount = *r.PendingTasksCount
		clusterInfo.ActiveServicesCount = *r.ActiveServicesCount

		clustersInfo = append(clustersInfo, clusterInfo)
	}

	return clustersInfo, err
}

// GetEcsClusterInstances - gets ECS cluster instances
func GetEcsClusterInstances(arn string) ([]string, error) {
	instances := []string{}

	svc := ecs.New(session.New())

	input := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(arn),
	}

	result, err := svc.ListContainerInstances(input)

	if err != nil {
		return instances, err
	}

	for _, instanceArn := range result.ContainerInstanceArns {
		re := regexp.MustCompile(`^arn:aws:ecs:.*:.*:container-instance/(.*)$`)
		m := re.FindStringSubmatch(*instanceArn)
		if len(m) > 0 {
			instances = append(instances, m[1])
		}
	}

	return instances, err
}

// GetEcsClusterInstancesInfo - gets ECS cluster instances info
func GetEcsClusterInstancesInfo(cluster string, instances []string) ([]EcsInstance, error) {
	instanceInfo := EcsInstance{}
	instancesInfo := []EcsInstance{}

	svc := ecs.New(session.New())

	input := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(cluster),
		ContainerInstances: aws.StringSlice(instances),
	}

	result, err := svc.DescribeContainerInstances(input)

	if err != nil {
		return instancesInfo, err
	}

	for _, ci := range result.ContainerInstances {
		pattern := `^arn:aws:ecs:.*:.*:container-instance/(.*)$`
		re := regexp.MustCompile(pattern)
		m := re.FindStringSubmatch(*ci.ContainerInstanceArn)
		if len(m) > 0 {
			instanceInfo.Name = m[1]
		} else {
			fmt.Printf("[ERROR] Instance ARN (%s) didn't match the pattern (%s)", *ci.ContainerInstanceArn, pattern)
		}

		instanceInfo.ARN = *ci.ContainerInstanceArn
		instanceInfo.Ec2InstanceID = *ci.Ec2InstanceId
		instanceInfo.Status = *ci.Status
		instanceInfo.RunningTasksCount = *ci.RunningTasksCount
		instanceInfo.PendingTasksCount = *ci.PendingTasksCount
		instanceInfo.AgentVersion = *ci.VersionInfo.AgentVersion
		instanceInfo.DockerVersion = strings.Replace(*ci.VersionInfo.DockerVersion, "DockerVersion: ", "", -1)

		for _, res := range ci.RemainingResources {
			if *res.Name == "CPU" {
				instanceInfo.RemainingCPU = *res.IntegerValue
			}
			if *res.Name == "MEMORY" {
				instanceInfo.RemainingMemory = *res.IntegerValue
			}

		}

		for _, att := range ci.Attributes {
			if *att.Name == "ecs.ami-id" {
				instanceInfo.AMI = *att.Value
			}
		}

		instancesInfo = append(instancesInfo, instanceInfo)
	}

	return instancesInfo, err
}

// GetEcsClusterArnByName - get ECS cluster ARN by cluster name
func GetEcsClusterArnByName(name string) (string, error) {
	arn := ""
	clusters, err := GetEcsClusters()

	if err != nil {
		return "", err
	}

	clustersInfo, err := GetEcsClustersInfo(clusters)

	if err != nil {
		return "", err
	}

	for _, clust := range clustersInfo {
		if clust.Name == name {
			arn = clust.ARN
		}
	}

	return arn, nil
}

// GetEcsClusterNameByArn - get ECS cluster name by cluster ARN
func GetEcsClusterNameByArn(arn string) (string, error) {
	name := ""
	re := regexp.MustCompile(`^arn:aws:ecs:.*:.*:cluster/(.*)$`)
	m := re.FindStringSubmatch(arn)
	if len(m) > 0 {
		name = m[1]
	} else {
		return name, errors.New("No match")
	}

	return name, nil
}

// IsEcsClusterReady - check if cluster is ready,
// all instances are in ACTIVE state and if mustHaveRunningTasks is specified,
// all instances must have at least one running task
func IsEcsClusterReady(arn string, mustHaveRunningTasks bool, numberOfZeroTasksInstances int) bool {
	// Get cluster info
	clusterInfo, err := GetEcsClustersInfo([]string{arn})
	if err != nil {
		fmt.Printf("Failed to get cluster info: %s\n", err)
		os.Exit(1)
	}

	// Get cluster instances list
	instances, err := GetEcsClusterInstances(arn)
	if err != nil {
		fmt.Printf("Failed to get cluster instances: %s\n", err)
		os.Exit(1)
	}

	zeroTasksInstanceCnt := 0

	// Get cluster instances info
	instancesInfo, err := GetEcsClusterInstancesInfo(clusterInfo[0].Name, instances)
	if err != nil {
		fmt.Printf("\nFailed to get cluster instances info: %s\n", err)
		os.Exit(1)
	}

	for _, inst := range instancesInfo {
		if inst.Status != "ACTIVE" {
			return false
		}

		if inst.RunningTasksCount < 1 {
			zeroTasksInstanceCnt++
		}
	}

	if zeroTasksInstanceCnt > numberOfZeroTasksInstances {
		return false
	}

	return true
}

// StopEcsTask - stop task
func StopEcsTask(cluster string, task string) (string, error) {
	svc := ecs.New(session.New())

	input := &ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Task:    aws.String(task),
		Reason:  aws.String("Stopped by ecs-manager tool"),
	}

	result, err := svc.StopTask(input)

	if err != nil {
		return "FAILED", err
	}

	return *result.Task.DesiredStatus, nil
}

// UpdateEcsContainerAgent - updates ECS container agent
func UpdateEcsContainerAgent(cluster string, instance string) (string, error) {
	svc := ecs.New(session.New())

	input := &ecs.UpdateContainerAgentInput{
		Cluster:           aws.String(cluster),
		ContainerInstance: aws.String(instance),
	}

	result, err := svc.UpdateContainerAgent(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == ecs.ErrCodeNoUpdateAvailableException {
				return "UP TO DATE", nil
			}
		}
		return "FAILED", err
	}

	return *result.ContainerInstance.AgentUpdateStatus, nil
}

// GetEcsInstanceTasks - get tasks running on instance
func GetEcsInstanceTasks(cluster string, instance string) ([]string, error) {
	tasks := []string{}
	svc := ecs.New(session.New())
	input := &ecs.ListTasksInput{
		Cluster:           aws.String(cluster),
		ContainerInstance: aws.String(instance),
		DesiredStatus:     aws.String("RUNNING"),
	}

	result, err := svc.ListTasks(input)

	if err != nil {
		return tasks, err
	}

	for _, taskArn := range result.TaskArns {
		re := regexp.MustCompile(`^arn:aws:ecs:.*:.*:task/(.*)$`)
		m := re.FindStringSubmatch(*taskArn)
		if len(m) > 0 {
			tasks = append(tasks, m[1])
		}
	}

	return tasks, nil
}

// ActivateEcsContainerInstance drains instance
func ActivateEcsContainerInstance(cluster string, instance string) (string, error) {
	svc := ecs.New(session.New())

	input := &ecs.UpdateContainerInstancesStateInput{
		Cluster:            aws.String(cluster),
		ContainerInstances: aws.StringSlice([]string{instance}),
		Status:             aws.String("ACTIVE"),
	}

	result, err := svc.UpdateContainerInstancesState(input)

	if err != nil {
		return "FAILED", err
	}

	return *result.ContainerInstances[0].Status, nil
}

// DrainEcsContainerInstance drains instance
func DrainEcsContainerInstance(cluster string, instance string) (string, error) {
	svc := ecs.New(session.New())

	input := &ecs.UpdateContainerInstancesStateInput{
		Cluster:            aws.String(cluster),
		ContainerInstances: aws.StringSlice([]string{instance}),
		Status:             aws.String("DRAINING"),
	}

	result, err := svc.UpdateContainerInstancesState(input)

	if err != nil {
		return "FAILED", err
	}

	return *result.ContainerInstances[0].Status, nil
}
