package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Global configuration holding runtime information
var config *Config

// PrintError checks if an error message exists and performs several actions.
// If an error is found, the message is printed out and the return value is true.
// Otherwise, if an error is not found, no action is taken and return value is false.
func PrintError(err error) bool {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return true
	}
	return false
}

// FatalError checks if an error message exists and if it does, will then exit,
// the program immediately returning an error code of 1.
func FatalError(err error) {
	if err != nil {
		fmt.Printf("Fatal: %s\n", err.Error())
		os.Exit(1)
	}
}

// CleanImages ...
func CleanImages(sess *session.Session, region *string, repc chan Report) {

	report := NewReport(*region)

	svc := ec2.New(sess, &aws.Config{
		Region: region,
	})

	// Describe images in this region that belong to the current account
	ownedImages, err := GetOwnedImages(svc)
	if report.AddError(err) {
		repc <- report
		return
	}

	// A slice of images that are to be deregistered
	var filteredImages ImageSlice

	{ // Filter out images that are older than date

		expiryDate := time.Now().AddDate(0, 0, -config.MaxAge)

		filteredImages = ownedImages.filter(func(image *ec2.Image) bool {
			creationDate, err := time.Parse(time.RFC3339Nano, *image.CreationDate)
			if report.AddError(err) {
				return false
			}
			if creationDate.After(expiryDate) {
				return true
			}
			return false
		})

	}

	{ // Filter out images that are in use by instances

		instancesPages, err := GetInstancesPages(svc)
		if report.AddError(err) {
			repc <- report
			return
		}

		filteredImages = filteredImages.filter(func(image *ec2.Image) bool {
			for _, page := range instancesPages {
				for _, reservation := range page.Reservations {
					for _, instance := range reservation.Instances {
						if *image.ImageId == *instance.ImageId {
							return true
						}
					}
				}
			}
			return false
		})

	}

	{ // Filter images that are present in launch configurations
		autoscalingSvc := autoscaling.New(sess)

		launchConfigPages, err := GetLaunchConfigurationPages(autoscalingSvc)
		if report.AddError(err) {
			repc <- report
			return
		}

		filteredImages = filteredImages.filter(func(image *ec2.Image) bool {
			for _, page := range launchConfigPages {
				for _, launchConfiguration := range page.LaunchConfigurations {
					if *image.ImageId == *launchConfiguration.ImageId {
						return true
					}
				}
			}
			return false
		})
	}

	report.RemovedImages = filteredImages

	if config.DryRun == false {
		for _, image := range filteredImages {
			fmt.Printf("Deregistering '%s'...\n", *image.ImageId)
			_, err := svc.DeregisterImage(&ec2.DeregisterImageInput{
				ImageId: image.ImageId,
			})
			report.AddError(err)
		}
	}

	repc <- report
	return
}

func main() {

	config = NewConfiguration()

	// Create a new AWS session object
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create a new EC2 service from session
	svc := ec2.New(sess)

	// Describe all available EC2 regions and store them
	regions, err := svc.DescribeRegions(nil)
	FatalError(err)

	// Create channels for goroutines to communicate through
	// resc, errc := make(chan string), make(chan error)
	reportc := make(chan Report)

	// Create a goroutine for every discovered region
	for _, region := range regions.Regions {
		go CleanImages(sess, region.RegionName, reportc)
	}

	// Block until all goroutines have completed
	for i := 0; i < len(regions.Regions); i++ {
		fmt.Printf("=============================\n")
		PrintReport(<-reportc)
	}
	fmt.Printf("=============================\n")

}
