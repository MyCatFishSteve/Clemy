package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ImageSlice ...
type ImageSlice []*ec2.Image

func (imageSlice *ImageSlice) filter(function func(*ec2.Image) bool) ImageSlice {
	var filteredList ImageSlice
	for _, image := range *imageSlice {
		if !function(image) {
			filteredList = append(filteredList, image)
		}
	}
	return filteredList
}
