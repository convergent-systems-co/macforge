# ADR 0008 — License: Apache 2.0

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

`GOALS.md` states "Open source first" and "Remain permanently free. No paywalls." A license must be chosen for the empty `LICENSE` file. The choice affects:

- Contributor reach (enterprise legal teams accept some licenses faster than others).
- Patent protection (relevant for a tool that operates in Apple's signing infrastructure).
- Compatibility with downstream users embedding MacForge in commercial pipelines.
- Alignment with the broader open-standards story.

## Decision

**Apache License 2.0.**

- The full license text goes in `/LICENSE`.
- A short `SPDX-License-Identifier: Apache-2.0` header goes at the top of every Go source file.
- The `NOTICE` file lists the project name and copyright holder; required attribution lives there.
- Contributions are accepted under Apache 2.0 by default — no separate CLA in v0.x. (Reconsider if enterprise contributions warrant one.)

Copyright holder: **Convergent Systems Co.** (the GitHub org owning the repo).

## Consequences

### Positive

- **Explicit patent grant.** Apache 2.0's §3 patent license is meaningful for a tool that operates in Apple's signing ecosystem — anyone contributing patent-encumbered code grants those patents to users.
- **Enterprise-friendly.** Apache 2.0 sails through corporate legal review faster than GPL family.
- **OSI-approved and FSF-free.** Hits the "open-source first" and "permanently free" goals without ambiguity.
- **Matches similar-class tools.** Kubernetes, Terraform (until 2023), most of CNCF, etc.

### Negative

- **Notice / attribution requirement.** Downstream consumers must include a copy of the license and the `NOTICE` file. This is mild and well-understood.
- **No copyleft.** Downstream forks can stay proprietary. Acceptable — we don't need to coerce participation, we want broad adoption.

### Neutral

- License compatibility: Apache 2.0 is compatible with GPLv3 (one-way) and MIT/BSD (both ways). Our dependencies must remain Apache-compatible; this is the standard `Code.md §6` license check.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Apache 2.0** | Patent grant, enterprise-friendly, OSI/FSF, broad adoption | Notice requirement | **Chosen** |
| MIT | Maximum permissiveness; shortest text | No patent grant — risky for tooling that touches Apple's signing patents | Rejected — patent grant matters here |
| BSD-3-Clause | Similar to MIT plus name-use restriction | Still no patent grant | Rejected — same reason as MIT |
| MPL 2.0 | File-level copyleft (weak); patent grant | More complex; less familiar; usually not chosen for tooling | Rejected — Apache is simpler for the same outcome |
| GPLv3 | Strong copyleft; full patent grant | Hostile to enterprise embedding; violates "open standard" frictionlessness | Rejected — wrong copyleft profile for a release tool |
| AGPLv3 | Network-use copyleft | Even more enterprise-hostile; irrelevant (MacForge is not a network service) | Rejected — wrong fit |
| BUSL 1.1 / SSPL | Source-available license with delayed-FOSS or non-commercial restrictions | Conflicts with "Remain permanently free. No paywalls." | Rejected — violates GOALS.md principles |

## Implementation

- `/LICENSE` — full Apache-2.0 text (unmodified from the Apache Foundation).
- `/NOTICE` — minimal:
  ```
  MacForge
  Copyright 2026 Convergent Systems Co.

  This product includes software developed by Convergent Systems Co.
  ```
- Per-file header on Go sources:
  ```go
  // SPDX-License-Identifier: Apache-2.0
  // Copyright 2026 Convergent Systems Co.
  ```
- `README.md` includes a `## License` section pointing at `LICENSE`.

## Links

- [Apache License 2.0 — Apache Software Foundation](https://www.apache.org/licenses/LICENSE-2.0)
- Reference: `Code.md §6, §8` (license tracking and copyright/IP)
