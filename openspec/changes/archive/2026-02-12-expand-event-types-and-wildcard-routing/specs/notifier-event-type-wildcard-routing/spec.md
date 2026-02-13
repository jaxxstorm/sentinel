## ADDED Requirements

### Requirement: Notifier route event type matching SHALL support wildcard selection
Sentinel SHALL accept `*` in `notifier.routes[].event_types` to match all emitted event types.

#### Scenario: Wildcard route matches all event types
- **WHEN** a route defines `event_types: ["*"]`
- **THEN** Sentinel treats any emitted event type as matching that route's event-type filter

### Requirement: Literal and wildcard event matching SHALL be deterministic
Sentinel SHALL apply deterministic event-type matching semantics for routes containing literals, wildcard, or both.

#### Scenario: Literal-only route filters by explicit event types
- **WHEN** a route defines explicit event types without `*`
- **THEN** only events whose type equals one of those literals match

#### Scenario: Mixed route with wildcard behaves as match-all
- **WHEN** a route contains `*` together with explicit event types
- **THEN** Sentinel treats the route as wildcard-matching all event types

### Requirement: Wildcard support SHALL be backward-compatible with existing routing
Adding wildcard matching SHALL NOT change behavior for configurations that only use explicit event type lists.

#### Scenario: Existing explicit route behavior remains unchanged
- **WHEN** Sentinel loads a route definition that does not include `*`
- **THEN** event routing behavior remains equivalent to prior explicit-match semantics
