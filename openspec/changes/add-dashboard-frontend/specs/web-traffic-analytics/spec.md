## MODIFIED Requirements

### Requirement: Aggregate reporting storage
The system SHALL maintain aggregate tables for common reporting dimensions and time grains.

#### Scenario: Dashboard reads extended aggregates
- **WHEN** a dashboard requests historical reports
- **THEN** the system SHALL read aggregate tables by default and return available PV, IP, visitor, session, event, and referrer metrics without scanning raw-event partitions
