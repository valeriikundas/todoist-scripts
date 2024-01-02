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
	"github.com/pkg/errors"
	"github.com/valeriikundas/todoist-scripts/api"
	"log"
	"os"
	"strings"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	region := "eu-central-1"
	config := aws.NewConfig().WithRegion(region).WithCredentialsChainVerboseErrors(true)
	session := session.Must(session.NewSession(config))
	secretsManager := secretsmanager.New(session, &aws.Config{})

	secretsOutput, err := secretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: jsii.String("gtd-secrets"),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "",
		}, errors.Wrap(err, "failed to get secret value")
	}

	var secrets struct {
		TodoistApiToken  string
		TelegramApiToken string
		TelegramUserID   string
	}
	err = json.Unmarshal([]byte(*secretsOutput.SecretString), &secrets)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "",
		}, errors.Wrap(err, "failed to unmarshal secret")
	}

	excludeFromZeroProjectsString, ok := os.LookupEnv("ExcludeFromZeroProjectsList")
	if !ok {
		return events.APIGatewayProxyResponse{StatusCode: 500,
			Body: "",
		}, errors.New("ExcludeFromZeroProjectsList environment variable is not set")
	}
	excludeFromZeroProjectsList := strings.Split(excludeFromZeroProjectsString, ";")

	resp, err := api.SendReportAboutIncorrectProjectsToTelegram(
		secrets.TodoistApiToken,
		secrets.TelegramApiToken,
		secrets.TelegramUserID,
		excludeFromZeroProjectsList,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "",
		}, errors.Wrap(err, "failed to get incorrect projects")
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "",
		}, errors.Wrap(err, "failed to marshal response")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(b),
	}, nil
}

func main() {
	runtimeApi := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	if runtimeApi == "" {
		resp, err := handler(context.TODO(), events.APIGatewayProxyRequest{})
		if err != nil {
			log.Fatal(err)
		}
		log.Print(resp)
		return
	}

	lambda.Start(handler)
}
