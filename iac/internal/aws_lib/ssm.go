package aws_lib

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/service/ssm"
)

type SSM struct {
	client *ssm.SSM
}

func NewSSM() *SSM {
	ssmClient := ssm.New(sess)
	return &SSM{
		client: ssmClient,
	}
}

func (ssmClient *SSM) GetParam(path string, isEncrypt bool) map[string]string {
	input := &ssm.GetParameterInput{
		Name:           &path,
		WithDecryption: &isEncrypt,
	}
	output, err := ssmClient.client.GetParameter(input)
	if err != nil {
		log.Fatal(err)
	}

	outParam := map[string]string{}
	json.Unmarshal([]byte(*output.Parameter.Value), &outParam)
	return outParam
}
