package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
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
