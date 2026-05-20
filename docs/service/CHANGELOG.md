# Service Manager Changelog

## [1.0.0] — 2026-05-20T04:27:00Z

### Added
- `Manager` interface with 7 methods (Enable, Disable, Start, Stop, Restart, Status, IsEnabled)
- `Status` enum type with human-readable String() method
- Build tag system for platform-specific implementations
- Factory function `NewManager()` for transparent platform selection
