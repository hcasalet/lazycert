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
    
    
# TODO for TE implementation
* ~~Update crypto.go to read from pem file when the file is already present.~~
* Implement Accept
  * Count incoming messages and create signed certificate for a block.
* Implement self-promotion
  * Count self promotion messages and broadcast leader status.
* TE LOG
  * Add log entries
  * Certificates

## Open questions
* How to find out *f*? 
  * Should we set it as a configurable option as part of the TE configuration?
  * Should we infer it at runtime based on registrations?
    * In this case how long should we wait for registrations to continue before selecting *f*?
