.PHONY: build clean deploy tf-apply-dev tf-apply-prod

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/auth src/services/authentication/auth/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/authorizer src/services/authentication/authorizer/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/ping src/services/authentication/ping/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/refreshToken src/services/authentication/refreshToken/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/logout src/services/authentication/logout/main.go
	
clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose

tf-apply-dev:
	cd terraform/environments/dev; terraform apply -auto-approve

tf-apply-prod:
	cd terraform/environments/prod; terraform apply -auto-approve
