package ecs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// Instance holds information about ECS instance
type Instance struct {
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
}

// Cluster holds information about ECS cluster
type Cluster struct {
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

// GetClusters - gets list of ECS clusters
func GetClusters() ([]string, error) {
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

// GetClustersInfo - gets ECS clusters info
func GetClustersInfo(arns []string) ([]Cluster, error) {
	clusterInfo := Cluster{}
	clustersInfo := []Cluster{}

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

// GetClusterInstances - gets ECS cluster instances
func GetClusterInstances(arn string) ([]string, error) {
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

// GetClusterInstancesInfo - gets ECS cluster instances info
func GetClusterInstancesInfo(cluster string, instances []string) ([]Instance, error) {
	instanceInfo := Instance{}
	instancesInfo := []Instance{}

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

		for _, att := range ci.Attributes {
			if *att.Name == "ecs.ami-id" {
				instanceInfo.AMI = *att.Value
			}
		}

		instancesInfo = append(instancesInfo, instanceInfo)
	}

	// fmt.Printf("Instances: %#v\n", instances)
	return instancesInfo, err
}

// StopTask - stop task
func StopTask(cluster string, task string) (string, error) {
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

// UpdateContainerAgent - updates ECS container agent
func UpdateContainerAgent(cluster string, instance string) (string, error) {
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

// GetInstanceTasks - get tasks running on instance
func GetInstanceTasks(cluster string, instance string) ([]string, error) {
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

// ActivateContainerInstance drains instance
func ActivateContainerInstance(cluster string, instance string) (string, error) {
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

// DrainContainerInstance drains instance
func DrainContainerInstance(cluster string, instance string) (string, error) {
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

// TerminateContainerInstance terminates instance
func TerminateContainerInstance(instance string) (string, error) {
	svc := ec2.New(session.New())

	input := &ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{instance}),
	}

	result, err := svc.TerminateInstances(input)

	if err != nil {
		return "FAILED", err
	}

	return strings.ToUpper(*result.TerminatingInstances[0].CurrentState.Name), nil
}
