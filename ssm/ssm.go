package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func GetParameterFromParameterStore(paramName string) (string, error) {
	// Create a new session
	sess, err := session.NewSession()
	if err != nil {
		return "", fmt.Errorf("unable to create session: %w", err)
	}

	// Create an SSM client from the session
	ssmClient := ssm.New(sess)

	// Get the parameter
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("ocurrió un error al consultar el parámetro: %w", err)
	}

	return *output.Parameter.Value, nil
}
