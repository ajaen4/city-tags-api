package containers

import (
	"city-tags-api-iac/internal/config"
	"context"
	"fmt"
	"log"

	registryClient "cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"github.com/pulumi/pulumi-docker-build/sdk/go/dockerbuild"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/api/iterator"
)

type Image struct {
	name       string
	repository *artifactregistry.Repository
	ctx        *pulumi.Context
	Resource   *dockerbuild.Image
}

func NewImage(ctx *pulumi.Context, cfg *config.ServiceCfg, name string, repository *artifactregistry.Repository) *Image {
	return &Image{
		name:       name,
		repository: repository,
		ctx:        ctx,
	}
}

func (img *Image) PushImage(version string) pulumi.StringInput {
	imageTag := fmt.Sprintf("%s:%s", img.name, version)

	imageURI := img.repository.Name.ApplyT(func(repoName string) string {
		return fmt.Sprintf("europe-west1-docker.pkg.dev/sityex-dev/%s/%s", repoName, imageTag)
	}).(pulumi.StringInput)

	push := img.repository.Name.ApplyT(func(repoName string) bool {
		push := !img.imageExists(repoName, version)
		return push
	}).(pulumi.BoolInput)

	var err error
	img.Resource, err = dockerbuild.NewImage(
		img.ctx,
		img.name,
		&dockerbuild.ImageArgs{
			Dockerfile: &dockerbuild.DockerfileArgs{
				Location: pulumi.String("../Dockerfile.api"),
			},
			Context: &dockerbuild.BuildContextArgs{
				Location: pulumi.String("../"),
			},
			Platforms: dockerbuild.PlatformArray{
				dockerbuild.Platform_Linux_amd64,
				dockerbuild.Platform_Linux_arm64,
			},
			Tags: pulumi.StringArray{imageURI},
			Push: pulumi.BoolInput(push),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return imageURI
}

func (img *Image) imageExists(repoName, version string) bool {
	ctx := context.Background()
	client, err := registryClient.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	parent := fmt.Sprintf(
		"projects/%s/locations/%s/repositories/%s",
		"sityex-dev",
		"europe-west1",
		repoName,
	)

	it := client.ListDockerImages(
		ctx,
		&artifactregistrypb.ListDockerImagesRequest{
			Parent: parent,
		},
	)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for _, tagInRepo := range resp.GetTags() {
			fmt.Println(tagInRepo)
			if tagInRepo == version {
				return true
			}
		}
	}
	return false
}
