# Process Management Changelog

## [1.0.0] — 2026-05-20T04:27:00Z

### Added
- mongod launch with --fork for daemonization
- Process kill via pgrep PID lookup and os.Process.Kill()
- Combined status detection (direct process + OS service)
- Launch history recording with timestamps
