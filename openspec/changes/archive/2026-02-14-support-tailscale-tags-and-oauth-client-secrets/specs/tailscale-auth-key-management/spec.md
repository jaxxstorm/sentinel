## MODIFIED Requirements

### Requirement: Sentinel SHALL enforce auth key precedence
Sentinel SHALL resolve auth key precedence as CLI flag over environment variable over configuration file, and SHALL treat an available auth key as higher precedence than OAuth credential onboarding sources.

#### Scenario: CLI flag overrides environment and config
- **WHEN** auth key values are present in flag, env, and config
- **THEN** Sentinel uses the flag value and ignores lower-precedence sources

#### Scenario: Auth key takes precedence over OAuth credentials
- **WHEN** Sentinel has both a resolved auth key and valid OAuth credential inputs
- **THEN** Sentinel uses auth-key onboarding and does not switch to OAuth mode
