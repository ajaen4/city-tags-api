package aws_lib

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type ECR struct {
	client *ecr.ECR
}

func NewECR() *ECR {
	ecrClient := ecr.New(sess)
	return &ECR{
		client: ecrClient,
	}
}

func (ecrClient ECR) IsImageInECR(repositoryURL, imageTag string) bool {
	splitURL := strings.Split(repositoryURL, "/")
	repoName := splitURL[len(splitURL)-1]
	images, err := ecrClient.client.ListImages(&ecr.ListImagesInput{
		RepositoryName: &repoName,
		Filter: &ecr.ListImagesFilter{
			TagStatus: aws.String("TAGGED"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, imageId := range images.ImageIds {
		if *imageId.ImageTag == imageTag {
			return true
		}
	}
	return false
}
