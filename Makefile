.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth-customers/auth src/services/authentication/auth/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
