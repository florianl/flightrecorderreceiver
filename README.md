# FlightRecorder Receiver

The FlightRecorder receiver collects profiles from files specified with a glob pattern.


## Configuration
- `include` (required): The glob path for files to watch
- `collection_interval` (default = `1m`): The interval at which metrics are emitted by this receiver.
- `initial_delay` (default = `1s`): defines how long this receiver waits before starting.

### Example

```
receivers:
  flightrecorder:
    include: /tmp/flightrecorder/*
    collection_interval: 10s
    initial_delay: 1s
```
