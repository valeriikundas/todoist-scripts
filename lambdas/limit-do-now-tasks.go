package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/jsii-runtime-go"
	"github.com/valeriikundas/todoist-scripts/api"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	region := "eu-central-1"
	config := aws.NewConfig().WithRegion(region)
	session := session.Must(session.NewSession(config))
	secretsManager := secretsmanager.New(session, &aws.Config{})

	todoistApiTokenSecretValueOutput, err := secretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: jsii.String("TodoistApiToken"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "",
		}, err
	}

	telegramApiTokenSecretValueOutput, err := secretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: jsii.String("TelegramApiToken"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "",
		}, err
	}

	telegramUserIDSecretValueOutput, err := secretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: jsii.String("TelegramUserID"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "",
		}, err
	}

	todoistApiToken := todoistApiTokenSecretValueOutput.String()
	telegramApiToken := telegramApiTokenSecretValueOutput.String()
	telegramUserID := telegramUserIDSecretValueOutput.String()

	resp, err := api.GetIncorrectProjects(todoistApiToken, telegramApiToken, telegramUserID)
	b, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(b),
	}, nil
}

func main() {
	lambda.Start(handler)
}
