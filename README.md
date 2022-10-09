# Bep2prom

Export Prometheus metrics from Bazel build events.

## Build

You can build via:

    $ bazel build //cmd/bep2prom
    INFO: Analyzed target //cmd/bep2prom:bep2prom (192 packages loaded, 10707 targets configured).
    INFO: Found 1 target...
    Target //cmd/bep2prom:bep2prom up-to-date:
      bazel-bin/cmd/bep2prom/bep2prom_/bep2prom
    INFO: Elapsed time: 95.844s, Critical Path: 29.88s
    INFO: 392 processes: 18 internal, 374 linux-sandbox.
    INFO: Build completed successfully, 392 total actions

You can copy `bazel-bin/cmd/bep2prom/bep2prom_/bep2prom` to your PATH:

    $ cp bazel-bin/cmd/bep2prom/bep2prom_/bep2prom /usr/local/bin/

## Run

Simple:

    $ bazel run //cmd/bep2prom
    INFO: Analyzed target //cmd/bep2prom:bep2prom (0 packages loaded, 0 targets configured).
    INFO: Found 1 target...
    Target //cmd/bep2prom:bep2prom up-to-date:
      bazel-bin/cmd/bep2prom/bep2prom_/bep2prom
    INFO: Elapsed time: 0.134s, Critical Path: 0.00s
    INFO: 1 process: 1 internal.
    INFO: Build completed successfully, 1 total action
    INFO: Build completed successfully, 1 total action
    2022/10/09 16:56:00 Listening on :8799

Then, when you run `bazel build` or `bazel test` for your build:

    $ bazel build --bes_backend=grpc://localhost:8799 //...

will send build events to `bep2prom`. Prometheus metrics will be exposed by `bep2prom` on `:2112` by default.
