# p16-t13: std.log — Structured Logging

## Purpose
Implement a structured logging library for AXIOM programs with log levels, context fields, configurable sinks (stderr, file, JSON), and zero-allocation fast paths for disabled log levels.

## Context
Production AXIOM programs need structured logging for observability. `std.log` provides level-filtered logging with key-value context, supporting both human-readable and machine-readable (JSON) output formats. Log calls at disabled levels must be zero-cost (no allocation, no string formatting).

## Inputs
- `std.fmt` Display interface from p16-t11
- `std.time` from p16-t10 for log timestamps
- `std.io` from p16-t04 for log output sinks

## Outputs
- `stdlib/log/log.ax` — Logger, LogRecord, log macros
- `stdlib/log/sink.ax` — Stderr, File, JSON sinks

## Dependencies
- p16-t11: std-fmt — Display for log values
- p16-t10: std-time — timestamps
- p16-t04: std-io — file sink I/O

## Detailed Requirements

```axiom
# stdlib/log/log.ax

type LogLevel: Trace, Debug, Info, Warn, Error, Fatal

type Logger:
    var level:  LogLevel
    var sink:   LogSink
    var fields: HashMap[str, str]  # context fields

    fn new(level: LogLevel, sink: LogSink) -> Logger
    fn with_field(self, key: str, val: str) -> Logger  # returns new Logger with field added

    fn trace(self, msg: str)
    fn debug(self, msg: str)
    fn info(self, msg: str)
    fn warn(self, msg: str)
    fn error(self, msg: str)
    fn fatal(self, msg: str)  # logs + exit(1)

# Global logger
var global_logger: Logger

fn set_global_logger(l: Logger)
fn trace(msg: str)
fn debug(msg: str)
fn info(msg: str)
fn warn(msg: str)
fn error(msg: str)

# stdlib/log/sink.ax
interface LogSink:
    fn write(self, record: LogRecord)

type LogRecord:
    var level:   LogLevel
    var time:    SystemTime
    var message: str
    var fields:  HashMap[str, str]

type StderrSink:     impl LogSink
type FileSink:       impl LogSink
type JsonSink:       impl LogSink  # outputs newline-delimited JSON
type MultiSink:      impl LogSink  # fan-out to multiple sinks
```

Zero-allocation fast path:
```axiom
fn debug(self, msg: str):
    if self.level > LogLevel::Debug:
        return   # no allocation if debug disabled
    self.sink.write(LogRecord{...})
```

JSON output format:
```json
{"time":"2026-05-16T10:30:00Z","level":"info","msg":"server started","port":"8080"}
```

## Implementation Steps

1. Create `stdlib/log/log.ax` — Logger with level filtering.
2. Implement global logger singleton with atomic swap.
3. Create `stdlib/log/sink.ax` — StderrSink, FileSink, JsonSink.
4. Implement zero-allocation level check.
5. Implement `with_field()` for context enrichment.
6. Write tests verifying level filtering and JSON output format.

## Test Plan
- `TestLogLevelFilter`: debug() at INFO level → no output
- `TestLogFields`: with_field("key", "val") → field appears in output
- `TestLogJSON`: JsonSink → valid JSON per line
- `TestLogFatal`: fatal() → exits with code 1
- `TestLogConcurrent`: 16 goroutines logging simultaneously → no interleaving

## Validation Checklist
- [ ] Disabled levels produce zero allocations
- [ ] JSON output is valid (parseable by std.json)
- [ ] fatal() calls exit(1) after logging
- [ ] Thread-safe: global_logger set/get atomic

## Acceptance Criteria
- 10M log records/sec discarded at disabled level (zero allocation)

## Definition of Done
- [ ] `stdlib/log/log.ax` implemented
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| String interpolation in log call allocates even when disabled | Use fn(Logger) -> str lazy evaluation; check level before calling |

## Future Follow-up Tasks
- OpenTelemetry trace ID propagation in log fields
- Log sampling (emit 1 in N for high-volume logs)
