package containers

import (
	"city-tags-api-iac/internal/input"
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
	project    string
	region     string
	Resource   *dockerbuild.Image
	imgCfg     input.ImgCfg
}

func NewImage(ctx *pulumi.Context, imgCfg input.ImgCfg, project string, region string, name string, repository *artifactregistry.Repository) *Image {
	return &Image{
		name:       name,
		repository: repository,
		ctx:        ctx,
		project:    project,
		region:     region,
		imgCfg:     imgCfg,
	}
}

func (img *Image) PushImage(version string) pulumi.StringInput {
	imageTag := fmt.Sprintf("%s:%s", img.name, version)

	imageURI := img.repository.Name.ApplyT(func(repoName string) string {
		return fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s", img.region, img.project, repoName, imageTag)
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
				Location: pulumi.String(fmt.Sprintf("../%s", img.imgCfg.Dockerfile)),
			},
			Context: &dockerbuild.BuildContextArgs{
				Location: pulumi.String(fmt.Sprintf("../%s", img.imgCfg.Context)),
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
		img.project,
		img.region,
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
