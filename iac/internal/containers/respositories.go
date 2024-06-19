package containers

import (
	"log"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecr"
)

type Repository struct {
	EcrRepository *ecr.Repository
	Ctx           *pulumi.Context
}

func NewRepository(ctx *pulumi.Context, name string) *Repository {
	repository, err := ecr.NewRepository(
		ctx,
		name,
		&ecr.RepositoryArgs{
			Name:               pulumi.String(name),
			ImageTagMutability: pulumi.String("IMMUTABLE"),
			ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
				ScanOnPush: pulumi.Bool(true),
			},
			ForceDelete: pulumi.Bool(true),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return &Repository{
		EcrRepository: repository,
		Ctx:           ctx,
	}
}
