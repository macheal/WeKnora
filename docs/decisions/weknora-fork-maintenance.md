# WeKnora Fork Maintenance Policy

- Status: Accepted
- Date: 2026-10-20
- Owner: Wang
- Fork: `macheal/WeKnora`
- Upstream: `Tencent/WeKnora`

## Decision

WeKnora is maintained as an independent fork. Production extensions are based on
the release deployed to Kubernetes, currently `v0.8.3`. Upstream contributions
are prepared on temporary branches based on the latest `upstream/main`; they are
not developed directly on the production extension branch.

## Branch Model

| Branch | Base | Purpose | Merge policy |
| --- | --- | --- | --- |
| `main` | `upstream/main` | Clean mirror of Tencent | Fast-forward only |
| `codex/weknora-extensions` | Release tag, currently `v0.8.3` | Long-lived production extensions | Merge feature branches; do not rebase |
| `feat/*`, `fix/*` | `codex/weknora-extensions` | Internal features and fixes | Merge into the extension branch |
| `upstream/feat/*`, `upstream/fix/*` | `upstream/main` | Minimal PR-ready changes | Rebase and submit to Tencent |

The production extension branch and upstream contribution branches are
separate. This keeps upstream `main` changes out of the deployed line until a
version upgrade has been tested.

## Version and Upgrade Rules

The deployed release is the compatibility baseline, not the latest untagged
`main` commit.

1. Deploy and test an upstream release before changing the extension baseline.
2. Upgrade in a temporary branch:
   `git switch -c chore/upgrade-vX.Y.Z codex/weknora-extensions`.
3. Merge the new upstream release tag or main snapshot into that branch.
4. Resolve conflicts in the upgrade branch and run the version-specific
   migration and regression checks.
5. Merge the upgrade branch into `codex/weknora-extensions`.
6. Tag the downstream release, for example `v0.8.3-macheal.1`; never overwrite
   an upstream tag.

`main` must stay a clean upstream mirror. Do not commit local extensions to
`main`.

## Upstream PR Flow

Before opening a PR against `Tencent/WeKnora:main`:

```bash
git fetch upstream --prune --tags
git switch -c upstream/feat/feature-name upstream/main
# Apply or cherry-pick only the upstream-ready commits.
git merge upstream/main
# Resolve conflicts locally and run targeted tests.
git push -u origin upstream/feat/feature-name
```

The PR target is `Tencent/WeKnora:main`. The PR must not contain internal
addresses, credentials, product-specific policy, or unrelated downstream
changes.

After the upstream PR is merged, sync the result into the extension line as a
separate migration change. Never assume an upstream merge automatically makes
the downstream extension branch production-ready.

## Compatibility Barriers

Prefer changes that are easy to carry across releases:

- Put new capabilities behind new files, packages, routes, or adapters.
- Isolate core changes and document the upstream file and commit they touch.
- Gate incomplete behavior behind a feature flag that defaults to off.
- Keep database migrations additive and test both PostgreSQL and SQLite paths.
- Record conflicts, migration steps, and verification in the upgrade PR.

## Required Verification

For an extension change, run the tests for the touched packages and the
frontend build when UI files change. For an upgrade, additionally run the
migration checks, memory-related regressions, and a smoke test of the deployed
configuration. Record any environment-dependent failure rather than omitting it.

## Current Baseline

- Production release: `v0.8.3`
- Release commit: `3e22822c1d2397be730c458c7539d0eb8c113a24`
- Production extension branch: `codex/weknora-extensions`
- Intended deployment: local Kubernetes
