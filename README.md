# same-bar-execution-reproduction

This is a study-owned Go module with an unsealed exploration lane. It begins
with no sources, frozen method, run, or result. The platform checkout is linked
through `go.mod`'s local `replace`; update that path if the repositories move.

## Start a lineage

1. Complete `charter.md`: question, study kind, target claim, non-goals, source
   permissions, and sources that cannot be shared with external AI services.
2. Put permitted source bytes in ignored `data/`. Do not commit them. Obtain
   and record actual acquisition details; never invent metadata or citations.
   For an untouched holdout, declare its planned artifact in the editable
   manifest with `availability: "post_freeze"`, `size: 0`, and `sha256: ""`.
   Its locator, release identity, acquisition status, checksum relationship,
   and retention policy are still required; its acquisition date stays empty in
   the frozen manifest. Record the actual acquisition date in study-owned
   status or provenance notes. Its declared path must remain absent through
   `check` and `freeze`. The run refuses to start until its nonempty local bytes
   have been acquired after freeze and hashed into that run's plan. That first
   run binds the path, size, and SHA-256 for every later run in the lineage. If
   a download must be corrected or replaced, create a new lineage rather than
   retrying with different bytes.
3. From this repository root, run the platform owner CLI:

   ```sh
   go run research-platform/cmd/research-platform manifest same-bar-execution-reproduction data/<file>
   go run research-platform/cmd/research-platform spec new same-bar-execution-reproduction --kind original
   ```

   Use `--kind reproduction` for a paper reproduction. Fill the manifest's
   acquisition fields and the spec's method, facts, choices, uncertainty,
   acceptance or comparison rule, and non-goals. After editing JSON in a text
   editor, run `go run research-platform/cmd/research-platform fmt same-bar-execution-reproduction`:
   this canonicalizes both draft records and refreshes the spec's manifest
   binding. `manifest` also refreshes that binding when source bytes change.
   `check <lineage>` reports all blockers. A second lineage needs
   `manifest <lineage> --study same-bar-execution-reproduction`.
4. Commit the completed draft records and study code. For original research,
   define a separate-data split in the spec before seeing holdout data. Then
   freeze from the Git root:

   ```sh
   go run research-platform/cmd/research-platform check same-bar-execution-reproduction
   go run research-platform/cmd/research-platform freeze same-bar-execution-reproduction --exploration none
   git add records && git commit -m 'Freeze lineage'
   ```

   Use `same-data` or `separate-data` and a one-line `--note` if there was
   exploration. A frozen lineage is immutable; a method change needs a new one.

## Run and publish

Implement only `study.Evaluate(in Inputs)` and `study.Verify(in Inputs, result)`
for the frozen method and study-owned result schema. Both receive the source
references bound to the run. Use `in.OpenSource(ref)` and check its read and
close errors; the platform checks size and SHA-256 at EOF or close. The CLI
calls `LoadLineage`, `BeginRun`, immutable
publication, sealing, offline integrity verification, and handoff generation:

```sh
go run ./cmd/same-bar-execution-reproduction run same-bar-execution-reproduction
go run ./cmd/same-bar-execution-reproduction verify <run-id>
go run ./cmd/same-bar-execution-reproduction handoff same-bar-execution-reproduction <run-id>
```

A run prints its generated ID. A study error leaves an `UNSEALED` run; retry
with a new run. `COMPLETE`, `INCOMPLETE`, and `TECHNICAL_FAILURE` have distinct
sealed outputs. `verify` checks both integrity and the study's semantic record.
`verify` also reports whether the current committed study implementation
matches the sealed one, or whether that comparison is unavailable for a dirty
tree. When it differs, the notice gives the sealed run's Git commit; use
`git worktree add <directory> <commit>` to replay with the recorded study code.
The handoff lists every run in the lineage, including `UNSEALED`; it refuses
corrupt sealed evidence. Commit `.research-platform/` and `handoffs/` after
sealing. The next run requires a clean Git tree, so commit evidence first.
Keep Git history and the study backup current.

## Exploration boundary

`explore/` may read local data and use any tools, including notebooks and AI
assistance. Label all its outputs exploratory. They are unsealed, never cited
in a handoff, and cannot establish a confirmatory claim. AI assistance is
external and untrusted: study code has no AI client. Send source bytes or
excerpts to an external AI service only when that source's license and
confidentiality permit it. Record sources that cannot be shared in the charter.

A confirmatory claim requires a frozen lineage whose freeze records whether
and how the data were explored. `same-data` is allowed, but its handoff names
that limitation. For original research, decide and record the separate-data
split before looking at the holdout.
