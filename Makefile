.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/auth src/services/authentication/auth/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/authorizer src/services/authentication/authorizer/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/ping src/services/authentication/ping/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/refreshToken src/services/authentication/refreshToken/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
