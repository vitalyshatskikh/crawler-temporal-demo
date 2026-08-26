## Purpose

Stores per-source crawling configuration, including an optional `update_interval_sec` field specifying how often documents from that source should be refreshed.

## ADDED Requirements

### Requirement: surf_configs update_interval_sec field

The `surf_configs` table SHALL include an `update_interval_sec INTEGER NOT NULL DEFAULT 86400` column.

#### Scenario: Default value on insert

- **WHEN** a `surf_configs` row is inserted without specifying `update_interval_sec`
- **THEN** the column value is 86400

#### Scenario: Custom value on insert

- **WHEN** a `surf_configs` row is inserted with `update_interval_sec = 600`
- **THEN** the stored value is 600

### Requirement: surfing.Params model requires update_interval_sec > 0

The `surf_params` type used in surfer workflows SHALL include an `update_interval_sec` integer field validated as greater than zero.

#### Scenario: Zero value rejected

- **WHEN** `surf_params` is constructed with `update_interval_sec = 0`
- **THEN** validation fails

#### Scenario: Negative value rejected

- **WHEN** `surf_params` is constructed with `update_interval_sec = -1`
- **THEN** validation fails

#### Scenario: Positive value accepted

- **WHEN** `surf_params` is constructed with `update_interval_sec = 7200`
- **THEN** validation succeeds

### Requirement: Seed data update_interval_sec values

The seeded `surf_configs` rows SHALL have the following `update_interval_sec` values:

| Config name | update_interval_sec |
|-------------|-------------------|
| siteapi-local-debug | 600 |
| siteapi-demo-fresh | 7200 |
| siteapi-demo-all | 7200 |

#### Scenario: Seed values loaded

- **WHEN** migrations are applied with seed data
- **THEN** the three seeded rows have the update intervals specified above
