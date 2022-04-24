API_PROTO_FILES=$(shell find api -name *.proto)

install:
	go install
	protoc -I=./api \
		   -I=./third_party \
		   --gogofaster_out=. \
		   --ginsev_out=. \
		   $(API_PROTO_FILES)

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@latest

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --gogofaster_out=. \
 	       --openapi_out==paths=source_relative:. \
 	       --swagger_out=logtostderr=true:. \
	       $(API_PROTO_FILES)

struct:
	protoc --proto_path=./api \
    	   --proto_path=./third_party \
     	   --gogofaster_out=source_relative:./api  \
     	   $(API_PROTO_FILES)

swagger:
	protoc --proto_path=./api \
           --proto_path=./third_party \
		   --swagger_out=logtostderr=true:. \
           $(API_PROTO_FILES)

doc:
	docker run -p 8081:8080 -e SWAGGER_JSON=/foo/api.swagger.json -v $(PWD):/foo swaggerapi/swagger-ui