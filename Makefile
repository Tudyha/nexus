.PHONY: proto

proto:
	cd pkg/proto && protoc --go_out=. --go_opt=paths=source_relative ./*.proto

.PHONY: build

build: 
	docker build -t nexus .
	docker tag nexus registry.cn-guangzhou.aliyuncs.com/knodio/nexus
	docker push registry.cn-guangzhou.aliyuncs.com/knodio/nexus