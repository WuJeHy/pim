gen_protocol_api:
	protoc  --go_out=plugins=grpc:.  api/pim.proto


show_log:
	tail -100f ./logs/pim.log