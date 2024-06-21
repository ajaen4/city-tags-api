package containers

import (
	"city-tags-api-iac/internal/aws_lib"
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

	push := pulumi.All(img.repository.RepositoryUrl, tag).ApplyT(
		func(args []any) bool {
			ecr := aws_lib.NewECR()
			push := !ecr.IsImageInECR(args[0].(string), args[1].(string))
			return push
		},
	).(pulumi.BoolInput)

	_, err := dockerbuild.NewImage(
		img.ctx,
		img.name,
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
			Push: push,
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
