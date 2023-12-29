package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
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
		jsii.String("gtd-secrets"),
		&awscdk.SecretsManagerSecretOptions{},
	)

	// todo: roles

	// lambdas
	limitDoNowTasksFunction := awscdklambdagoalpha.NewGoFunction(
		stack,
		jsii.String("limit-do-now-tasks"),
		&awscdklambdagoalpha.GoFunctionProps{
			Runtime:      awslambda.Runtime_GO_1_X(),
			Entry:        jsii.String("lambdas/limit-do-now-tasks.go"),
			LogRetention: awslogs.RetentionDays_THREE_DAYS,
			//todo: rewrite with this one when it works
			//InitialPolicy: &awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{}).
		},
	)

	//secrets := awssecretsmanager.Secret_FromSecretNameV2(stack, jsii.String("gtd-secrets"), jsii.String("gtd-secrets"))

	limitDoNowTasksFunction.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings("secretsmanager:GetSecretValue"),
		Resources: jsii.Strings(
			"***REMOVED***",

			//"***REMOVED***",
			//"arn:aws:secretsmanager:eu-central-1:***REMOVED***:secret:gtd-secrets"
			//*secrets.SecretArn(),
		),
	}))
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
			Env: &awscdk.Environment{
				Account: jsii.String("***REMOVED***"),
				Region:  jsii.String("eu-central-1"),
			},
		},
	})

	app.Synth(nil)
}
