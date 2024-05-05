SRC_DIR = .
DST_DIR = .

.PHONY:
proto:
	mkdir -p $(DST_DIR)/api
	mkdir -p $(DST_DIR)/types
	protoc --proto_path=$(SRC_DIR) --go_out=${DST_DIR}/types --go_opt=paths=source_relative --go-grpc_out=${DST_DIR}/types --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional felight_types.proto
	protoc --proto_path=$(SRC_DIR) --go_out=${DST_DIR}/api --go_opt=paths=source_relative --go-grpc_out=${DST_DIR}/api --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional felight_api.proto

.PHONY:
clean:
	rm -rf $(DST_DIR)/api
	rm -rf $(DST_DIR)/types