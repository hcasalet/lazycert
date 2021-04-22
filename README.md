# lazycert
Lazy Certification is a methodology which can be easily applied to consensus protocols such as Paxos to help make Paxos Byzantine fault tolerant.


# Command to generate go code from .proto file
    cd dump
    protoc --go_out=plugins=grpc:./lc lc.proto
# Command to build and run TE and test TE

### Run TE Server
    go run te.go
    
### Test with demo client
    go run te_client_demo.go 
