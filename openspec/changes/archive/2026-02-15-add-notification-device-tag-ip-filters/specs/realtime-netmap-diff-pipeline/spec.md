## MODIFIED Requirements

### Requirement: Sentinel SHALL track additional normalized change dimensions in realtime mode
Sentinel SHALL normalize and track additional field groups required for expanded event emission, including peer routes, peer tags, selected device identity attributes, normalized owner and peer IP identity values, and selected tailnet metadata.

#### Scenario: Route change affects normalized tracked dimensions
- **WHEN** a realtime update changes only a tracked route or tag field
- **THEN** Sentinel treats the update as meaningful for diff/event processing even when presence is unchanged

#### Scenario: Device owner or IP identity change affects normalized tracked dimensions
- **WHEN** a realtime update changes a normalized owner or peer IP identity value used for route selector matching
- **THEN** Sentinel treats the update as meaningful for diff/event processing even when presence is unchanged
