package main

import (
	"fmt"
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
	"github.com/valeriikundas/todoist-scripts/utils"
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
			Environment:   utils.ReadConfig(),
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

func main() {
	defer jsii.Close()

	tryReadDotenv()

	app := awscdk.NewApp(&awscdk.AppProps{})

	awsAccountID := getEnvVarOrPanic("AWS_ACCOUNT_ID")
	awsRegion := getEnvVarOrPanic("AWS_REGION")
	NewCdkStack(app, "gtd-scripts", &CdkStackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(awsAccountID),
				Region:  jsii.String(awsRegion),
			},
			Tags: &map[string]*string{
				"AppManagerCFNStackKey": jsii.String("gtd-scripts"),
			},
		},
	})

	app.Synth(nil)
}

func tryReadDotenv() {
	_, err := os.Stat(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return
		} else {
			must(err)
		}
	} else {
		// .env exists
		err = godotenv.Load(".env")
		must(err)
	}
}

func getEnvVarOrPanic(k string) string {
	val, ok := os.LookupEnv(k)
	if !ok {
		panic(fmt.Sprintf("%s environment variable is missing", k))
	}
	return val
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
