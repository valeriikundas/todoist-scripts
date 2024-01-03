package main

import (
	"encoding/json"
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
	"strings"
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

	// secrets
	secretsARN := os.Getenv("AWS_SECRETS_FULL_ARN")
	_ = awssecretsmanager.Secret_FromSecretCompleteArn(
		stack,
		jsii.String("gtd-secrets"),
		jsii.String(secretsARN),
	)

	// policy statements
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
			Entry:         jsii.String("lambdas/limit-do-now-tasks/main.go"),
			Runtime:       awslambda.Runtime_GO_1_X(),
			InitialPolicy: &[]awsiam.PolicyStatement{readSecretsPolicyStatement},
			Environment:   readEnvVars("config.json"),
		},
	)
	archiveOlderInboxTasks := awscdklambdagoalpha.NewGoFunction(
		stack,
		jsii.String("archive-older-inbox-tasks"),
		&awscdklambdagoalpha.GoFunctionProps{
			LogRetention:  awslogs.RetentionDays_THREE_DAYS,
			Timeout:       awscdk.Duration_Seconds(jsii.Number(30)),
			Entry:         jsii.String("lambdas/archive-older-inbox-tasks/main.go"),
			Runtime:       awslambda.Runtime_GO_1_X(),
			InitialPolicy: &[]awsiam.PolicyStatement{readSecretsPolicyStatement},
		},
	)

	// scheduling
	scheduleDaily8AM := awsevents.Schedule_Cron(&awsevents.CronOptions{
		Hour:   jsii.String("6"),
		Minute: jsii.String("0"),
	})
	_ = awsevents.NewRule(
		stack,
		jsii.String("run-limit-do-now-tasks-at-8am-daily"),
		&awsevents.RuleProps{
			Schedule: scheduleDaily8AM,
			Targets: &[]awsevents.IRuleTarget{
				awseventstargets.NewLambdaFunction(
					limitDoNowTasksFunction,
					&awseventstargets.LambdaFunctionProps{},
				),
			},
		},
	)

	awsevents.NewRule(stack, jsii.String("archive-older-inbox-tasks-daily"), &awsevents.RuleProps{
		Schedule: scheduleDaily8AM,
		Targets: &[]awsevents.IRuleTarget{
			awseventstargets.NewLambdaFunction(
				archiveOlderInboxTasks,
				&awseventstargets.LambdaFunctionProps{},
			),
		},
	})

	return stack
}

func readEnvVars(configFileName string) *map[string]*string {
	file, err := os.Open(configFileName)
	must(err)

	decoder := json.NewDecoder(file)
	var config struct {
		ExcludeFromZeroProjectsList []string
	}
	err = decoder.Decode(&config)
	must(err)

	zeroProjectsListJoined := strings.Join(config.ExcludeFromZeroProjectsList, ";")
	envVars := &map[string]*string{
		"ExcludeFromZeroProjectsList": jsii.String(zeroProjectsListJoined),
	}
	return envVars
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	defer jsii.Close()

	err := godotenv.Load(".env")
	must(err)

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
