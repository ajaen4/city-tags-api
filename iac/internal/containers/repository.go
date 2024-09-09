package containers

import (
	"log"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewRepository(ctx *pulumi.Context, name string, region string) *artifactregistry.Repository {
	repository, err := artifactregistry.NewRepository(
		ctx,
		name,
		&artifactregistry.RepositoryArgs{
			Location:     pulumi.String(region),
			RepositoryId: pulumi.String(name),
			Description:  pulumi.String("test repos"),
			Format:       pulumi.String("DOCKER"),
		})
	if err != nil {
		log.Fatal(err)
	}

	return repository
}
