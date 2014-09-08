GO=		go
GOX=		gox

GO_PACKAGE=	github.com/fumiyas/qrc/cmd/qrc
CROSS_TARGETS=	linux darwin windows

default: build

get:
	$(GO) get

build:
	$(GO) build cmd/qrc/qrc.go

cross:
	$(GOX) -os="$(CROSS_TARGETS)" $(GO_PACKAGE)

