list:
	@echo "sync-codes: 同步代码"
	@echo "build: 编译可执行文件"
	@echo "test: 测试"

sync-codes:
	git pull

build:
	PATH=/tmp/govendor/bin:$(PATH)
	GOPATH=/tmp/govendor/:$(GOPATH)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o server ./main.go

deps:
	mkdir -p /tmp/govendor/src/gitlab.wallstcn.com/infrastructure/
	cp -r /builds/infrastructure/k8s-injector/vendor/* /tmp/govendor/src

	mkdir -p /tmp/govendor/bin
	cp -R "/builds/infrastructure/k8s-injector" "/tmp/govendor/src/gitlab.wallstcn.com/infrastructure/"
	rm -rf "/builds/infrastructure/k8s-injector/vendor"
	rm -rf "/tmp/govendor/src/gitlab.wallstcn.com/infrastructure/k8s-injector/vendor"

test:
	go test
