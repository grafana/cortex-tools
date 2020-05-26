# parselogs

A cli tool to parse Cortex query-frontend logs and format them for easy analysis.

```
Options:
  -dur duration
        only show queries which took longer than this duration, e.g. -dur 10s
  -query
        show the query
  -utc
        show timestamp in UTC time
```

Feed logs into it using `logcli` from Loki, `kubectl` from Kubernetes, `cat` from a file, or any other way to get raw logs:

```
$ cat query-frontend-logs.log | ./parselogs -dur 5s
Timestamp                                TraceID           Length    Duration       Status  Path
2020-05-26 13:51:15.0577354 -0400 EDT    76b9939fd5c78b8f  6h0m0s    10.249149614s  (200)   /api/prom/api/v1/query_range
2020-05-26 13:52:15.771988849 -0400 EDT  2e7473ab10160630  10h33m0s  7.472855362s   (200)   /api/prom/api/v1/query_range
2020-05-26 13:53:46.712563497 -0400 EDT  761f3221dcdd85de  10h33m0s  11.874296689s  (200)   /api/prom/api/v1/query_range
```
