
all: container_comms internal_comms

common: common.proto
	protoc -I=. --go_out=../common --go_opt=paths=source_relative \
	--go-grpc_out=../common --go-grpc_opt=paths=source_relative \
	common.proto

container_comms: volpe_container.proto common
	protoc -I=. --go_out=../container --go_opt=paths=source_relative \
	--go-grpc_out=../container --go-grpc_opt=paths=source_relative \
	volpe_container.proto

	python3 -m grpc_tools.protoc -I. --python_out=../pybindings --pyi_out=../pybindings \
	--grpc_python_out=../pybindings volpe_container.proto

internal_comms: volpe_internal.proto common
	protoc -I=. --go_out=../internal --go_opt=paths=source_relative \
	--go-grpc_out=../internal --go-grpc_opt=paths=source_relative \
	volpe_internal.proto