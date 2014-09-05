GO=		go

REPO=		github.com/fumiyas/qrc

default: build

build:
	$(GO) build cmd/qrc/qrc.go

get:
	$(GO) get
