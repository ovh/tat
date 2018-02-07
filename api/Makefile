ECHO = @echo
GO = go

help:
	$(ECHO) "TaT API Makefile"
	$(ECHO) "'make go-build'	Build the API, don't forget to source your custom configuration in environment variables"

go-build:
	$(GO) build -ldflags "-X github.com/ovh/tat.Version=$(VERSION)"
