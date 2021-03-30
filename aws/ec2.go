package aws

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// TerminateEc2Instance terminates instance
func TerminateEc2Instance(instance string) (string, error) {
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

// IsEc2InstanceTerminated - check if instance is terminated
func IsEc2InstanceTerminated(instance string) bool {
	svc := ec2.New(session.New())

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instance),
		},
	}

	result, err := svc.DescribeInstances(input)

	if err != nil {
		fmt.Printf("Failed to get instance info: %s\n", err)
		os.Exit(1)
	}

	if *result.Reservations[0].Instances[0].State.Name == "terminated" {
		return true
	}

	return false
}
