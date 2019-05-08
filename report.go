package main

import "fmt"

// Report will hold information about actions taken on regions and if any errors occurred.
type Report struct {
	Region        string
	RemovedImages ImageSlice
	Errors        []error
}

// NewReport will return a new report object with the region assigned.
func NewReport(region string) (report Report) {
	return Report{
		Region: region,
	}
}

// AddError will add an error to the report and return true if error not nil.
func (r *Report) AddError(err error) bool {
	if err != nil {
		r.Errors = append(r.Errors, err)
		return true
	}
	return false
}

// PrintReport takes a report object and prints the information in an appropriate format.
func PrintReport(report Report) {

	fmt.Printf("Region: %s\n", report.Region)

	// Print image information
	if len(report.RemovedImages) > 0 {
		if !config.DryRun {
			fmt.Printf("Images removed:\n")
		} else {
			fmt.Printf("Images to be removed:\n")
		}
		for idx, image := range report.RemovedImages {
			fmt.Printf("%3d: %s\n", idx, *image.ImageId)
		}
	} else {
		fmt.Printf("No images were removed\n")
	}

	if len(report.Errors) > 0 {
		fmt.Printf("Errors encountered:\n")
		for idx, error := range report.Errors {
			fmt.Printf("%d: %s\n", idx, error)
		}
	}

}
