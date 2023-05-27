# create:
# 	protoc -I=proto --go_out=gen/ proto/*.proto
# 	protoc --go-grpc_out=gen/ proto/*.proto -I=proto
# 	# protoc --dart_out=grpc:../frontend/lib/pb/ -Iproto proto/*.proto
# 	protoc -I . --grpc-gateway_out ./gen \
#     --grpc-gateway_opt logtostderr=true \
#     --grpc-gateway_opt paths=source_relative \
#     --grpc-gateway_opt generate_unbound_methods=true \
#     proto/service.proto

generate:
	cd proto; buf generate

clean:
	rm -rf pkg

git:
	git add .
	git commit -a -m "$m"
	git push -u origin main


	# https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > google/api/http.proto
	# https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > google/api/annotations.proto