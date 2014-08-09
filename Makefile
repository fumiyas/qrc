GO=		GOPATH="$$PWD" go

REPO=		github.com/fumiyas/qrc

default: build

build:
	$(GO) build

get:
	$(GO) get
