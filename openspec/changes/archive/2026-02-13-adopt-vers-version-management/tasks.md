## 1. Version metadata foundation

- [x] 1.1 Add `github.com/jaxxstorm/vers` to module dependencies.
- [x] 1.2 Introduce an internal version metadata adapter/package that exposes version, commit, and build timestamp to callers.
- [x] 1.3 Define documented fallback values for untagged/local builds in the adapter.

## 2. CLI version command behavior

- [x] 2.1 Refactor `internal/cli/version.go` to read version data from the new adapter instead of direct constants.
- [x] 2.2 Ensure `sentinel version` remains config-independent and does not initialize tsnet or runtime services.
- [x] 2.3 Verify stable output fields for release and local builds (version, commit, build timestamp).

## 3. Build and release metadata wiring

- [x] 3.1 Update build metadata injection (GoReleaser/ldflags or equivalent adapter inputs) to match the vers-backed model.
- [x] 3.2 Ensure semantic git tag builds propagate tag/version metadata into runtime version output.
- [x] 3.3 Validate CI/release workflow compatibility after metadata wiring changes.

## 4. Tests and regression coverage

- [x] 4.1 Add unit tests for version adapter behavior with tagged metadata inputs.
- [x] 4.2 Add unit tests for fallback metadata behavior when tag/build values are absent.
- [x] 4.3 Add/adjust CLI tests for `sentinel version` output expectations.

## 5. Documentation and operator guidance

- [x] 5.1 Update docs to describe version metadata source and expected `sentinel version` output semantics.
- [x] 5.2 Update release documentation to reflect how version metadata is injected for tagged binaries.
- [x] 5.3 Run targeted tests and `go test ./...` to validate end-to-end behavior before archive.
