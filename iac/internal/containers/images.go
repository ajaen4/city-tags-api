package containers

import (
	"fmt"
	"log"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
	"github.com/pulumi/pulumi-docker-build/sdk/go/dockerbuild"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Image struct {
	name       string
	repository *ecr.Repository
	ctx        *pulumi.Context
}

func NewImage(ctx *pulumi.Context, name string, repository *ecr.Repository) *Image {
	return &Image{
		name:       name,
		repository: repository,
		ctx:        ctx,
	}
}

func (img *Image) PushImage(version string) pulumi.StringInput {
	authToken := img.authenticate()
	tag := img.repository.RepositoryUrl.ApplyT(func(repositoryUrl string) string {
		return fmt.Sprintf("%s:%s-%s", repositoryUrl, img.name, version)
	}).(pulumi.StringInput)

	_, err := dockerbuild.NewImage(
		img.ctx,
		fmt.Sprintf("%s-image", img.name),
		&dockerbuild.ImageArgs{
			Dockerfile: &dockerbuild.DockerfileArgs{
				Location: pulumi.String("../Dockerfile"),
			},
			Context: &dockerbuild.BuildContextArgs{
				Location: pulumi.String("../"),
			},
			Platforms: dockerbuild.PlatformArray{
				dockerbuild.Platform_Linux_amd64,
				dockerbuild.Platform_Linux_arm64,
			},
			Registries: dockerbuild.RegistryArray{
				&dockerbuild.RegistryArgs{
					Address:  img.repository.RepositoryUrl,
					Password: authToken.Password(),
					Username: authToken.UserName(),
				},
			},
			Tags: pulumi.StringArray{tag},
			Push: pulumi.Bool(true),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return tag
}

func (img *Image) authenticate() ecr.GetAuthorizationTokenResultOutput {
	return ecr.GetAuthorizationTokenOutput(
		img.ctx,
		ecr.GetAuthorizationTokenOutputArgs{
			RegistryId: img.repository.RegistryId,
		},
		nil,
	)
}
