module github.com/valeriikundas/todoist-scripts/encore

go 1.21.4

replace github.com/valeriikundas/todoist-scripts v0.0.0-20231130152949-6426a3e7a29f => ../

require (
	encore.dev v1.27.0
	github.com/valeriikundas/todoist-scripts v0.0.0-20231130152949-6426a3e7a29f
)

require github.com/google/uuid v1.4.0 // indirect
