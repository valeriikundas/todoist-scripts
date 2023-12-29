build-lambda:
	GOOS=linux GOARCH=amd64 go build -C lambdas -o limit-do-now-tasks .
