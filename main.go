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
var config = NewConfiguration()

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

// GetOwnedImages returns an ImageSlice that contains images owned by the account of the service.
func GetOwnedImages(svc *ec2.EC2) (ImageSlice, error) {
	images, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String("self"),
		},
	})
	return images.Images, err
}

// CleanImages ...
func CleanImages(sess *session.Session, region *string, resc chan string, errc chan error) {

	svc := ec2.New(sess, &aws.Config{
		Region: region,
	})

	// Describe images in this region that belong to the current account
	ownedImages, err := GetOwnedImages(svc)
	if err != nil {
		errc <- err
		return
	}

	// Images to be deleted
	var filteredImages ImageSlice

	{ // Filter out images that are in use by instances
		var instancesOutputSlice []*ec2.DescribeInstancesOutput
		err := svc.DescribeInstancesPages(nil, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			instancesOutputSlice = append(instancesOutputSlice, page)
			return true
		})
		if err != nil {
			errc <- err
			return
		}
		filteredImages = ownedImages.filter(func(image *ec2.Image) bool {
			for _, page := range instancesOutputSlice {
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

	{ // Filter out images that are older than date
		expiryDate := time.Now().AddDate(0, 0, -14)
		filteredImages = filteredImages.filter(func(image *ec2.Image) bool {
			creationDate, err := time.Parse(time.RFC3339Nano, *image.CreationDate)
			if PrintError(err) {
				return false
			}
			if creationDate.After(expiryDate) {
				return true
			}
			return false
		})
	}

	{ // Filter images that are present in launch configurations
		autoscalingSvc := autoscaling.New(sess)

		var launchConfigurationSlice []*autoscaling.DescribeLaunchConfigurationsOutput
		err := autoscalingSvc.DescribeLaunchConfigurationsPages(nil, func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
			launchConfigurationSlice = append(launchConfigurationSlice, page)
			return true
		})
		if err != nil {
			errc <- err
			return
		}
		filteredImages = filteredImages.filter(func(image *ec2.Image) bool {
			for _, page := range launchConfigurationSlice {
				for _, launchConfiguration := range page.LaunchConfigurations {
					if *image.ImageId == *launchConfiguration.ImageId {
						return true
					}
				}
			}
			return false
		})
	}

	if config.DryRun == false {
		for _, image := range filteredImages {
			fmt.Printf("Deregistering '%s'...\n", *image.ImageId)
			_, err := svc.DeregisterImage(&ec2.DeregisterImageInput{
				ImageId: image.ImageId,
			})
			PrintError(err)
		}
	}

	resc <- *region + " completed"
	return
}

func main() {

	// Print config information
	if config.DryRun {
		fmt.Println("Dry run enabled, program is running non-destructively")
	}

	// Create a new AWS session object
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create a new EC2 service from session
	svc := ec2.New(sess)

	// Describe all available EC2 regions and store them
	regions, err := svc.DescribeRegions(nil)
	if PrintError(err) {
		os.Exit(1)
	}

	// Create channels for goroutines to communicate through
	resc, errc := make(chan string), make(chan error)

	// Create a goroutine for every discovered region
	for _, region := range regions.Regions {
		go CleanImages(sess, region.RegionName, resc, errc)
	}

	// Block until all goroutines have completed
	for i := 0; i < len(regions.Regions); i++ {
		select {
		case res := <-resc:
			if config.Verbose {
				defer fmt.Println(res)
			}
		case err := <-errc:
			PrintError(err)
		}
	}

}
