package containers

import (
	"log"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/artifactregistry"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewRepository(ctx *pulumi.Context, name string) *artifactregistry.Repository {
	repository, err := artifactregistry.NewRepository(
		ctx,
		name,
		&artifactregistry.RepositoryArgs{
			Location:     pulumi.String("europe-west1"),
			RepositoryId: pulumi.String(name),
			Description:  pulumi.String("test repos"),
			Format:       pulumi.String("DOCKER"),
		})
	if err != nil {
		log.Fatal(err)
	}

	return repository
}
