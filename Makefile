UNAME := $(shell uname)
build: install
	GOOS=darwin go build -o bin/ghrelease_osx ghrelease.go
	GOOS=linux go build -o bin/ghrelease_lin ghrelease.go
	GOOS=windows go build -o bin/ghrelease_win ghrelease.go
	chmod +x bin/*

release: build
ifeq ($(UNAME),Linux)
	echo "Detected Linux" 
	./bin/ghrelease_lin -settings=release.json
endif
ifeq ($(UNAME),Darwin)
	echo "Detected OSx"
	./bin/ghrelease_osx -settings=release.json
endif

install:
	go get golang.org/x/oauth2
	go get github.com/google/go-github/github
	go get github.com/sirupsen/logrus

test:
	go test . -v
