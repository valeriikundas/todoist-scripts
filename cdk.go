package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
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
	_ = awscdk.SecretValue_SecretsManager(
		jsii.String("todoistApiToken"),
		&awscdk.SecretsManagerSecretOptions{},
	)
	_ = awscdk.SecretValue_SecretsManager(
		jsii.String("telegramApiToken"),
		&awscdk.SecretsManagerSecretOptions{},
	)
	_ = awscdk.SecretValue_SecretsManager(
		jsii.String("telegramUserID"),
		&awscdk.SecretsManagerSecretOptions{},
	)

	// todo: roles

	// lambdas
	limitDoNowTasksFunction := awslambda.NewFunction(
		stack,
		jsii.String("limit-do-now-tasks"),
		&awslambda.FunctionProps{
			Code: awslambda.Code_FromAsset(
				jsii.String("./api/"),
				&awss3assets.AssetOptions{},
			),
			Handler: jsii.String("lambda/limit-do-now-tasks.go"),
			Runtime: awslambda.Runtime_GO_1_X(),
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

	app := awscdk.NewApp(nil)

	NewCdkStack(app, "gtd-scripts", &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
