---
title: Disabled Features
sort_rank: 10
---

# Disabled Features

Here is a list of features that are disabled by default since they are breaking changes or are considered experimental.
Their behaviour can change in future releases which will be communicated via the [release changelog](https://github.com/prometheus/prometheus/blob/master/CHANGELOG.md).

You can enable them using the `--enable-feature` flag with a comma separated list of features.
They may be enabled by default in future versions.

## `@` Modifier in PromQL

`--enable-feature=promql-at-modifier`

The `@` modifier lets you specify the evaluation time for instant vector selectors,
range vector selectors, and subqueries. More details can be found [here](querying/basics.md#-modifier).

## Remote Write Receiver

`--enable-feature=remote-write-receiver`

The remote write receiver allows Prometheus to accept remote write requests from other Prometheus servers. More details can be found [here](storage.md#overview).

## In Memory Exemplar Storage

`--enable-feature=exemplar-storage`

This stores the exemplars exposed via OpenMetrics format into a circular queue in the memory. Use `--storage.exemplars.exemplars-limit` to set the limit on number of exemplars.
More details on querying exemplars can be found [here](querying/api.md#querying-exemplars).
