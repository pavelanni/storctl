# Using limactl

## Status

Accepted

## Date

2024-01-29

## Context

- Need to work with virtual machines on a macOS or Linux host
- Lima is a tool of choice on macOS and also available on Linux
- Lima can use QEMU as a backend virtualization engine
- Lima is written in Go
- There is an option to use Lima's Go library or use the CLI tool `limactl` from inside our Go code
- There is also Colima that is a wrapper for Lima but it uses Docker instead of `containerd` as a containerization engine

## Decision

Use `limactl` CLI calls from inside Go code.

1. The Go library used by Lima developers is not created for external consumption.
   To use it we would have to go deep into its ecosystem and import a lot of other dependencies.

2. The CLI interface is more stable so we are more safe in this case.

3. In our case, performance is not a significant factor, so there is not much gain in using the native Go library.

## Consequences

### Positive

- More stable interface
- Fewer imports/dependencies

### Negative

- Need to parse the command JSON output
- Slightly lower performance (but not visible by the end user)

## Related Decisions


## Notes

