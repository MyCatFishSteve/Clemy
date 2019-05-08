package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestImageSlice(t *testing.T) {

	var imageSlice ImageSlice = []*ec2.Image{
		&ec2.Image{
			Name: aws.String("Machine-1"),
		},
		&ec2.Image{
			Name: aws.String("Machine-2"),
		},
		&ec2.Image{
			Name: aws.String("Machine-3"),
		},
		&ec2.Image{
			Name: aws.String("Machine-4"),
		},
	}

	filtered := imageSlice.filter(func(arg1 *ec2.Image) bool {
		if *arg1.Name == "Machine-1" || *arg1.Name == "Machine-2" {
			return false
		}
		return true
	})

	if len(filtered) != 2 {
		t.Error("Invalid filtered data length")
	}

	if *filtered[0].Name != "Machine-1" || *filtered[1].Name != "Machine-2" {
		t.Error("Filter data error")
	}

}
