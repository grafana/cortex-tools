1. go to `./cmd/benchtool`
2. execute commands
```shell
    go build
```
3. forward port from dev cluster using this command:
```shell
  kubectl port-forward service/gateway -n ge-metrics 8081:80
```
4. review `./workload.yml` file and change name of the metric or count of labels if you want 

5. replace `tenan-id` with your tenant id, `token` with your access token and run the command:
```shell
   ./benchtool -bench.instance-name="tenant-id" -bench.write.enabled=true -bench.write.basic-auth-username="tenant-id" -bench.write.basic-auth-password="token" -bench.write.endpoint="localhost:8081" -bench.ring-check.enabled=false -bench.workload-file-path="./workload.yml"
```
NOTE: it's necessary to create a token with scope METRICS_WRITE and replace token provided in example 
