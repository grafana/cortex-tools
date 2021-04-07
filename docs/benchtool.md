# Benchtool

The `benchtool` is a small load testing utility to generate load for Prometheus
remote-write and query API endpoints. It uses a statically configured YAML
workload file to describe the series and queries used for the generate load.

## Workload file

The workload file can be configured as a follows:

```yaml
replicas: <int>
queries:
  - expr_template: <string>
    interval: <duration>
    num_queries: <int>
    series_type: <enum>
series:
  - name: <string>
    type: <enum>
    static_labels:
      "<string>": "<string>"
    labels:
      - name: <string>
        unique_values: <int>
        value_prefix: <string>
write_options:
  batch_size: <int>
  interval: <duration>
```

```yaml
    queries:
      - expr_template: sum(<<.Name>>{<<.Matchers>>})
        interval: 1m
        num_queries: 0
        series_type: gauge-random
      - expr_template: sum(<<.Name>>{<<.Matchers>>})
        interval: 1m
        num_queries: 0
        series_type: gauge-zero
        time_range: 2h
      - expr_template: sum(rate(<<.Name>>{<<.Matchers>>}[1m]))
        interval: 1m
        num_queries: 0
        series_type: counter-random
    replicas: 0
    series:
      - labels:
          - name: workload_label_01
            unique_values: 2
            value_prefix: workload_label_01
          - name: workload_label_02
            unique_values: 5
            value_prefix: workload_label_02
          - name: workload_label_03
            unique_values: 100
            value_prefix: workload_label_03
        name: workload_metric_gauge_zero_00001
        static_labels:
            static: "true"
        type: gauge-zero
    write_options:
        batch_size: 1000
        interval: 15s
```
