GO=		go
GO_PACKAGE=	github.com/fumiyas/qrc/cmd/qrc

CROSS_TARGETS=	linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

default: build

.PHONY: default build test vet fmt cross clean

build:
	$(GO) build ./cmd/qrc

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

cross:
	@for target in $(CROSS_TARGETS); do \
		os=$${target%%/*}; arch=$${target##*/}; \
		ext=; [ "$$os" = "windows" ] && ext=.exe; \
		out=qrc-$$os-$$arch$$ext; \
		echo "==> $$out"; \
		GOOS=$$os GOARCH=$$arch $(GO) build -o $$out $(GO_PACKAGE) || exit $$?; \
	done

clean:
	rm -f qrc qrc-*-* qrc-*-*.exe
