package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Config ...
type Config struct {
	dryRun bool
}

// IsImageActive returns a boolean value based on if an image ID is currently in use
func IsImageActive(svc *ec2.EC2, imageID string) (bool, error) {
	res, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("image-id"),
				Values: []*string{
					aws.String(imageID),
				},
			},
		},
	})
	if err != nil {
		return false, err
	}
	// If there is any reservation, then the image is in use
	if len(res.Reservations) > 0 {
		fmt.Println("Active reservation found for " + imageID)
		return true, nil
	}
	return false, nil
}

// GetOwnedImages returns a slice of ec2.Image types that belong to the calling service.
func GetOwnedImages(svc *ec2.EC2) ([]*ec2.Image, error) {
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

	// Store the current time exactly 14 days ago
	expiryDate := time.Now().AddDate(0, 0, -14)

	// Find candidate images for removal
	for _, ownedImage := range ownedImages {

		creationTime, _ := time.Parse(time.RFC3339Nano, *ownedImage.CreationDate) // Parse the CreationDate into a Time object
		if creationTime.After(expiryDate) {                                       // The image is still younger than expiry date, skip to next image
			continue
		}

		active, err := IsImageActive(svc, *ownedImage.ImageId) // Check if image is currently being used by an instance
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if active {
			continue
		}

		// Send a request to deregister the image
		fmt.Println(fmt.Sprintf("Delete candidate found: %s. (%s)", *ownedImage.ImageId, *region))
		// svc.DeregisterImage(&ec2.DeregisterImageInput{
		// 	ImageId: imageCandidate.ImageId,
		// })
	} // End of image candidate loop

	resc <- "Region completed (" + *region + ")"
	return
}

func main() {

	// Create a new AWS session object
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create a new EC2 service from session
	svc := ec2.New(sess)

	// Describe all available EC2 regions and store them
	regions, _ := svc.DescribeRegions(nil)

	fmt.Println("Discovered regions:", len(regions.Regions))

	// Create channels for goroutines to communicate through
	resc, errc := make(chan string), make(chan error)

	// Create a goroutine for every discovered region
	fmt.Println("Cleaning AMIs in all regions")
	for _, region := range regions.Regions {
		go CleanImages(sess, region.RegionName, resc, errc)
	}

	// Block until all goroutines have completed
	for i := 0; i < len(regions.Regions); i++ {
		select {
		case res := <-resc:
			fmt.Println(res)
		case err := <-errc:
			fmt.Println(err.Error())
		}
	}

}
