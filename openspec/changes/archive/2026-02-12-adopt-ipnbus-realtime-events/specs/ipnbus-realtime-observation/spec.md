## ADDED Requirements

### Requirement: Sentinel SHALL consume IPNBus notifications as the primary runtime source
Sentinel SHALL subscribe to the local Tailscale IPNBus using a watch stream and use bus notifications as the primary trigger for runtime change processing.

#### Scenario: Runtime establishes watch stream on startup
- **WHEN** Sentinel starts in normal run mode
- **THEN** Sentinel opens an IPNBus watch stream against the active local tsnet/localapi context before entering steady-state observation

### Requirement: Sentinel SHALL bootstrap source state from initial bus snapshots
Sentinel SHALL request initial state and initial netmap data from the watch stream so the first runtime baseline is derived from bus-provided state.

#### Scenario: Initial stream frame includes baseline netmap
- **WHEN** Sentinel receives initial watch notifications
- **THEN** Sentinel constructs its first observed netmap baseline from bus data without requiring a separate polling read

### Requirement: Sentinel SHALL recover from watch-stream interruptions
Sentinel SHALL reconnect the watch stream with bounded exponential backoff after transient stream failures.

#### Scenario: Stream disconnect triggers reconnect loop
- **WHEN** the IPNBus watch stream returns an error while runtime is still active
- **THEN** Sentinel retries stream establishment with backoff until reconnected or process shutdown

### Requirement: Sentinel SHALL preserve configuration compatibility for source behavior
Sentinel SHALL keep existing configuration keys valid while providing realtime observation as the default source behavior.

#### Scenario: Existing config remains valid under realtime source
- **WHEN** Sentinel runs with an existing configuration that previously used poll-driven observation
- **THEN** Sentinel starts successfully and uses realtime IPNBus observation without requiring breaking config changes
