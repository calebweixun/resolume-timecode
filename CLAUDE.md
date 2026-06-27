# Claude Project Context

Read and follow `AGENTS.md`; it is the canonical architecture, development,
release, and project-state guide for this repository.

Current handoff summary:

- Release tag `v1.0.2` points to release-fix commit `583026a`.
- `v1.0.2` built successfully as macOS Universal and Windows amd64.
- Its GitHub Release is still a draft and requires manual publication.
- macOS output is a ZIP containing `resolume-timecode.app`, not a DMG, and it
  is not signed or notarized.
- Before committing, check the worktree and run the validations listed in
  `AGENTS.md`.
