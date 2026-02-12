## MODIFIED Requirements

### Requirement: Human output defaults to pretty formatted mode
Sentinel SHALL default to human-oriented pretty output for CLI status and diff presentation using deterministic formatting rules, and SHALL route embedded Tailscale runtime logs through the same formatter instead of raw stdlib output lines.

#### Scenario: Default output is pretty
- **WHEN** an operator runs `sentinel diff` without specifying log format
- **THEN** Sentinel renders a styled human-readable diff summary

#### Scenario: Embedded runtime lines follow pretty formatter
- **WHEN** Sentinel runs in default pretty mode and tsnet/tailscaled emits user-visible runtime messages
- **THEN** those messages are emitted through Sentinel's pretty logger formatting instead of raw `YYYY/MM/DD` stdlib log lines

### Requirement: Sentinel supports structured JSON logging
Sentinel SHALL provide a JSON log format option that emits structured fields suitable for machine ingestion, and each runtime record SHALL include a stable `log_source` field identifying the emitting subsystem.

#### Scenario: JSON mode emits structured records
- **WHEN** Sentinel runs with `--log-format=json`
- **THEN** each log record is a valid JSON object with stable keys including level and timestamp

#### Scenario: Runtime records include source attribution
- **WHEN** Sentinel emits runtime log records in JSON mode
- **THEN** each record includes `log_source` with values such as `sentinel`, `tailscale`, or `sink`

### Requirement: NO_COLOR and no-color flag disable ANSI styling
Sentinel SHALL disable ANSI color/styling when `NO_COLOR` is set or `--no-color` is enabled.

#### Scenario: No-color output strips escape sequences
- **WHEN** `NO_COLOR=1 sentinel status` is executed
- **THEN** output contains no ANSI escape codes while preserving message content

### Requirement: Sensitive metadata is redacted in default logs
Sentinel SHALL avoid logging sensitive peer metadata by default and MUST support redaction rules in both pretty and JSON log modes.

#### Scenario: Sensitive field is redacted
- **WHEN** a peer metadata field is marked sensitive by redaction rules
- **THEN** Sentinel logs a redacted placeholder instead of the raw value
