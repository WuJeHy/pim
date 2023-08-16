gen_protocol_api:
	protoc  --go_out=. --go_opt=paths=source_relative api/pim.proto


show_log:
	tail -100f ./logs/pim.log