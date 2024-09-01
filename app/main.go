package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"app/ecrmanager"
	"app/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// Define command-line flags
	days := flag.Int("days", 365, "Number of days to evaluate against lastPulledTime.")
	region := flag.String("region", "us-east-1", "AWS region where the ECR is running.")
	mode := flag.String("mode", "ecr", "Mode of operation: 'ecr' for ECR cleanup, 'k8s' for Kubernetes pod checking.")

	// Customize the usage function to include flag descriptions
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Parse the flags; automatically handles -h/--help
	flag.Parse()

	// Validate the mode flag
	if *mode != "ecr" && *mode != "k8s" {
		log.Fatalf("Invalid mode: %s. Must be 'ecr' or 'k8s'.", *mode)
	}

	// AWS session setup with specified region
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*region),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}
	ecrSvc := ecr.New(sess)

	// Calculate the date threshold based on the days flag
	now := time.Now()
	dateOlderThan := now.AddDate(0, 0, -*days)

	// Kubernetes client setup if mode is 'k8s'
	var clientset *kubernetes.Clientset
	if *mode == "k8s" {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Failed to create in-cluster config: %v", err)
		}
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatalf("Failed to create Kubernetes client: %v", err)
		}
	}

	// Collect pods if running in Kubernetes mode
	var pods []v1.Pod
	if *mode == "k8s" {
		podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Failed to list pods: %v", err)
		}
		pods = podList.Items
	}

	var totalSizeNeverPulled, totalSizeOlderThan int64

	// List all ECR repositories
	repos, err := ecrSvc.DescribeRepositories(&ecr.DescribeRepositoriesInput{})
	if err != nil {
		log.Fatalf("Failed to describe repositories: %v", err)
	}

	// Iterate over all repositories
	for _, repo := range repos.Repositories {
		// List all images in the repository
		images, err := ecrSvc.ListImages(&ecr.ListImagesInput{
			RepositoryName: repo.RepositoryName,
		})
		if err != nil {
			log.Fatalf("Failed to list images: %v", err)
		}

		// Iterate over all images
		for _, imageID := range images.ImageIds {
			describeImagesOutput, err := ecrSvc.DescribeImages(&ecr.DescribeImagesInput{
				RepositoryName: repo.RepositoryName,
				ImageIds:       []*ecr.ImageIdentifier{imageID},
			})
			if err != nil {
				log.Fatalf("Failed to describe images: %v", err)
			}

			for _, imageDetail := range describeImagesOutput.ImageDetails {
				// Skip if the image is still being used by any pod in the cluster (for mode 'k8s')
				if *mode == "k8s" && ecrmanager.IsImageUsedByPods(*imageDetail.ImageDigest, pods) {
					continue
				}

				lastPullTime := imageDetail.LastRecordedPullTime
				imageSize := *imageDetail.ImageSizeInBytes

				if lastPullTime == nil {
					totalSizeNeverPulled += imageSize
					logger.LogJSON("INFO", fmt.Sprintf("Image %s (%s) is never pulled, size: %d", *imageDetail.ImageDigest, *repo.RepositoryName, imageSize))

					ecrmanager.DeleteImage(ecrSvc, repo.RepositoryName, imageID)
				} else if lastPullTime.Before(dateOlderThan) {
					totalSizeOlderThan += imageSize
					logger.LogJSON("INFO", fmt.Sprintf("Image %s (%s) was last pulled over %d days ago, size: %d", *imageDetail.ImageDigest, *repo.RepositoryName, *days, imageSize))

					ecrmanager.DeleteImage(ecrSvc, repo.RepositoryName, imageID)
				}
			}
		}
	}

	// Print total sizes in GB
	log.Printf("Total size of images never pulled: %.2f GB", float64(totalSizeNeverPulled)/1073741824)
	log.Printf("Total size of images older than %d days: %.2f GB", *days, float64(totalSizeOlderThan)/1073741824)
}
