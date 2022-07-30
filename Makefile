PROTOC = protoc
PACKAGE = github.com/ZentriaMC/vaultadm/pkg/proto

bufs += pkg/proto/vaultadm.pb.go

# ---------------------------------------------------
%.pb.go : %.proto
	$(PROTOC) --go_out=. --go_opt=paths=source_relative --go_opt=M$(<)=$(PACKAGE) \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=M$(<)=$(PACKAGE) \
		$<

protobufs: $(bufs)

clean:
	rm -f pkg/proto/*.pb.go

all: protobufs
