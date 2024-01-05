package lambdacommon

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/valeriikundas/todoist-scripts/utils"
	"log"
	"os"
)

func Run[R any](f func(secrets *Secrets) (*R, error)) {
	runWithEnv(withHandler(f))
}

func withHandler[R any](f func(secrets *Secrets) (*R, error)) func(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context,
		request events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		return withSetup(f)
	}
}

func runWithEnv(f func(ctx context.Context, request events.APIGatewayProxyRequest) (
	events.APIGatewayProxyResponse,
	error,
)) {
	runtimeApi := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	if runtimeApi == "" {
		config := *utils.ReadConfig()
		for key, value := range config {
			os.Setenv(key, *value)
		}

		err := godotenv.Load(".env")
		must(err)

		resp, err := f(context.TODO(), events.APIGatewayProxyRequest{})
		must(err)
		log.Print(resp)
		return
	}

	lambda.Start(f)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func withSetup[R any](f func(secrets *Secrets) (R, error)) (
	events.APIGatewayProxyResponse,
	error,
) {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	secrets, err := readSecrets()
	if err != nil {
		return empty500Response(), errors.Wrap(err, "failed to read secrets")
	}

	resp, err := f(secrets)
	if err != nil {
		return empty500Response(), err
	}

	return marshall[R](resp)
}

func empty500Response() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: 500,
		Body: "",
	}
}

const region = "eu-central-1"

type Secrets struct {
	TodoistApiToken  string
	TelegramApiToken string
	TelegramUserID   string
}

func readSecrets() (*Secrets, error) {
	config := aws.NewConfig().WithRegion(region).WithCredentialsChainVerboseErrors(true)
	sess := session.Must(session.NewSession(config))
	secretsManager := secretsmanager.New(sess, &aws.Config{})

	secretsOutput, err := secretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: jsii.String("gtd-secrets"),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get secret value")
	}

	var secrets *Secrets
	err = json.Unmarshal([]byte(*secretsOutput.SecretString), &secrets)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal secret")
	}

	return secrets, nil
}

func marshall[R any](resp R) (events.APIGatewayProxyResponse, error) {
	b, err := json.Marshal(resp)
	if err != nil {
		return empty500Response(), errors.Wrap(err, "failed to marshal response")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(b),
	}, nil
}
