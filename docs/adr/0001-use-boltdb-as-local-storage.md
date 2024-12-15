# Use of BBolt as Local Storage

## Status

Accepted

## Date

2024-12-12

## Context

- Getting information from the cloud takes time
- Need for local storage of lab configurations and state
- Requirement for embedded database
- Need for atomic operations
- No requirement for concurrent access from multiple processes

## Decision

Use BBolt (github.com/etcd-io/bbolt) as the embedded key-value store

## Consequences

### Positive

- Simple, reliable embedded storage
- ACID compliant
- No external dependencies
- File-based, easy to backup

### Negative

- Limited to single process access
- No built-in replication
