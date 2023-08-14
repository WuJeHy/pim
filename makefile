gen_protocol_api:
	protoc  --go_out=plugins=grpc:.  api/pim.proto