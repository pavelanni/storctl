# Kubernetes-style YAML Resource Definitions

## Status

Accepted

## Date

2024-12-12

## Context

- Need for declarative resource definitions
- Requirement for familiar format
- Need for versioning and metadata

## Decision

Use Kubernetes-style YAML format for resource definitions

## Consequences

### Positive

- Familiar to users of Kubernetes
- Built-in versioning
- Consistent metadata structure
- Easy to validate
- Separation between Spec and Status sections

### Negative

- May be overkill for simple resources
- Learning curve for non-Kubernetes users
