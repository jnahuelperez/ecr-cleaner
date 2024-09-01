package ecrmanager

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/ecr"
	v1 "k8s.io/api/core/v1"
)

// Function to check if any pod is using the image
func IsImageUsedByPods(imageID string, pods []v1.Pod) bool {
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			// Extract the ImageID from the container's image reference
			parts := strings.Split(container.Image, "@")
			if len(parts) == 2 {
				// Compare the image ID from the pod with the image ID from ECR
				if parts[1] == imageID {
					return true
				}
			}
		}
	}
	return false
}

// Function to delete the image (commented out)
func DeleteImage(ecrSvc *ecr.ECR, repoName *string, imageID *ecr.ImageIdentifier) {
	// Delete the image
	// _, err := ecrSvc.BatchDeleteImage(&ecr.BatchDeleteImageInput{
	// 	RepositoryName: repoName,
	// 	ImageIds:       []*ecr.ImageIdentifier{imageID},
	// })
	// if err != nil {
	// 	fmt.Printf("Failed to delete image: %v", err)
	// }
	// fmt.Printf("The image %s would be deleted\n", *imageID.ImageDigest)
}
