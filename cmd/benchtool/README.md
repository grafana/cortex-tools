1. go to this folder in terminal
2. execute commands
```shell
    go build
   ./benchtool -bench.instance-name="team-a" -bench.write.enabled=true -bench.write.basic-auth-username="team-a" -bench.write.basic-auth-password="dmxhZC10ZXN0LWFsbDpdLG0yKTdkPT4kNjk4ezA1NSE2VC5kMn4=" -bench.write.endpoint="localhost:8081" -bench.ring-check.enabled=false -bench.workload-file-path="./workload.yml"
```
NOTE: it's necessary to create a token with scope METRICS_WRITE and replace token provided in example 
