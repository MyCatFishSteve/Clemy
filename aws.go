package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	// InstancesPages is a slice of ec2.DescribeInstancesOutput pointers
	InstancesPages []*ec2.DescribeInstancesOutput
	// LaunchConfigurationPages is a slice of autoscaling.DescribeLaunchConfigurationsOutput pointers
	LaunchConfigurationPages []*autoscaling.DescribeLaunchConfigurationsOutput
)

// GetOwnedImages returns an ImageSlice that contains images owned by the account of the service.
func GetOwnedImages(svc *ec2.EC2) (ImageSlice, error) {
	images, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String("self"),
		},
	})
	return images.Images, err
}

// GetInstancesPages will return InstancesPages and an error if encountered.
func GetInstancesPages(svc *ec2.EC2) (pages InstancesPages, err error) {
	err = svc.DescribeInstancesPages(nil, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		pages = append(pages, page)
		return true
	})
	return
}

// GetLaunchConfigurationPages will return LaunchConfigurationPages and an error if encountered.
func GetLaunchConfigurationPages(svc *autoscaling.AutoScaling) (pages LaunchConfigurationPages, err error) {
	err = svc.DescribeLaunchConfigurationsPages(nil, func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		pages = append(pages, page)
		return true
	})
	return
}
