GO=		go
GOX=		gox

GO_PACKAGE=	github.com/fumiyas/qrc/cmd/qrc
CROSS_TARGETS=	linux/amd64 linux/386 darwin/amd64 windows/386

default: build

get:
	$(GO) get

build:
	$(GO) build cmd/qrc/qrc.go

cross:
	$(GOX) -osarch="$(CROSS_TARGETS)" $(GO_PACKAGE)

