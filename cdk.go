package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
	"os"
)

type CdkStackProps struct {
	awscdk.StackProps
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var stackProps awscdk.StackProps
	if props != nil {
		stackProps = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &stackProps)

	secretsARN := os.Getenv("AWS_SECRETS_FULL_ARN")
	_ = awssecretsmanager.Secret_FromSecretCompleteArn(
		stack,
		jsii.String("gtd-secrets"),
		jsii.String(secretsARN),
	)

	readSecretsPolicyStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("secretsmanager:GetSecretValue"),
		Resources: jsii.Strings(secretsARN),
	})

	// lambdas
	limitDoNowTasksFunction := awscdklambdagoalpha.NewGoFunction(
		stack,
		jsii.String("limit-do-now-tasks"),
		&awscdklambdagoalpha.GoFunctionProps{
			LogRetention:  awslogs.RetentionDays_THREE_DAYS,
			Timeout:       awscdk.Duration_Seconds(jsii.Number(30)),
			Entry:         jsii.String("lambdas/limit-do-now-tasks.go"),
			Runtime:       awslambda.Runtime_GO_1_X(),
			InitialPolicy: &[]awsiam.PolicyStatement{readSecretsPolicyStatement},
		},
	)

	// todo: write other lambda functions in cdk

	// scheduling
	_ = awsevents.NewRule(
		stack,
		jsii.String("run-limit-do-now-tasks-at-8am-daily"),
		&awsevents.RuleProps{
			Description: nil,
			Schedule: awsevents.Schedule_Cron(&awsevents.CronOptions{
				Hour:   jsii.String("8"),
				Minute: jsii.String("0"),
			}),
			Targets: &[]awsevents.IRuleTarget{
				awseventstargets.NewLambdaFunction(
					limitDoNowTasksFunction,
					&awseventstargets.LambdaFunctionProps{},
				),
			},
		},
	)
	// todo: add cdk for other lambda functions

	return stack
}

func main() {
	defer jsii.Close()

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	app := awscdk.NewApp(&awscdk.AppProps{})

	NewCdkStack(app, "gtd-scripts", &CdkStackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(os.Getenv("AWS_ACCOUNT_ID")),
				Region:  jsii.String(os.Getenv("AWS_REGION")),
			},
		},
	})

	app.Synth(nil)
}
