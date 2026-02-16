## 1. Installation Path Documentation

- [x] 1.1 Add GitHub Releases installation steps to docs (asset selection, download, extract, executable setup, and `sentinel version` verification).
- [x] 1.2 Add Docker-first runtime guidance for quick-start flows and cross-link to compose/deployment docs.
- [x] 1.3 Add a short run-path selection note (release binary vs Docker vs source development) in getting-started content.

## 2. README and Operator Example Alignment

- [x] 2.1 Rewrite README quick-start to prioritize release binary and Docker commands over `go run`.
- [x] 2.2 Keep `go run` examples only in clearly labeled development sections of README/docs.
- [x] 2.3 Update high-traffic operator pages (`docs/getting-started.md`, `docs/commands.md`, related release/deployment docs) to use consistent command style.

## 3. Consistency and Validation

- [x] 3.1 Verify links between README, configuration/getting-started docs, release docs, and docker docs are correct.
- [x] 3.2 Validate docs examples for command correctness and naming consistency with current release artifacts.
- [x] 3.3 Run docs preview or markdown lint checks used by the project and resolve any formatting or link issues introduced by this change.
