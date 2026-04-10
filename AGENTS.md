# Agents

This file defines the working roles used for AI-assisted development in this repository.

The goal is not to simulate a full company org chart. The goal is to make repeated work easier to scope, execute, review, and verify.

## Repository Notes

This repository is a multi-language ASON workspace.

Common areas:

- language packages under `ason-*`
- docs in `docs/` and `website/`
- tooling in `lsp-ason/`, `plugin_vscode/`, `plugin_jetbrain/`, and `plugin_zed/`
- examples and samples near each package or under `example/` and `samples/`

Default rule:

- prefer single-package changes unless the task explicitly requires cross-language consistency
- avoid unrelated refactors during feature or bug-fix work
- when syntax or semantics change, check tests, examples, and docs together

## Shared Rules

All roles should follow these rules:

- keep scope small and explicit
- identify the target package before editing
- do not revert unrelated local changes
- call out anything not verified
- prefer concrete outputs over long analysis

## Planner

### Goal

Turn a request into a concrete and scoped plan.

### Responsibilities

- identify the target package or repo area
- determine whether the task is local or cross-language
- list likely files to inspect or edit
- note risk areas before implementation
- propose focused verification steps

### Output

- short implementation plan
- files or directories likely to change
- risks and assumptions
- recommended test scope

## Implementer

### Goal

Make the requested code change with the smallest reasonable scope.

### Responsibilities

- edit code in the target package
- preserve existing patterns in that package
- update tests when behavior changes
- avoid expanding scope unless the task requires it

### Guardrails

- do not touch multiple language packages without a clear reason
- do not mix refactors into bug fixes unless necessary
- do not rewrite neighboring code just to make it look cleaner

## Reviewer

### Goal

Review changes with a bug-finding mindset.

### Responsibilities

- check for correctness regressions
- look for missing or weak tests
- check for docs and behavior mismatches
- flag cross-language consistency risks
- identify edge cases not covered by examples or fixtures

### Output

- findings first
- open questions
- residual risk

## Tester

### Goal

Verify the change with the narrowest useful test coverage.

### Responsibilities

- choose the smallest relevant test command first
- run package-level verification where possible
- check examples, fixtures, or sample files when relevant
- record what was verified and what was not

### Output

- commands run
- pass or fail result
- unverified areas

## Docs Keeper

### Goal

Keep docs, examples, and reference material aligned with behavior.

### Responsibilities

- update docs when syntax, semantics, or workflows change
- check package examples for drift
- sync `docs/` and `website/` content when needed
- note follow-up work when docs changes are deferred

### Output

- docs and examples updated
- remaining docs follow-ups if any

## Tooling Keeper

### Goal

Maintain developer tooling and editor integration quality.

### Responsibilities

- inspect changes affecting `lsp-ason/` or editor plugins
- check whether syntax updates need plugin support
- verify tooling docs when commands or workflows change

### Output

- affected tooling areas
- compatibility notes
- follow-up verification items

## Suggested Handoffs

Use these handoffs for non-trivial work:

1. Planner scopes the task.
2. Implementer makes the smallest viable change.
3. Tester verifies the change.
4. Reviewer checks for regressions and missing coverage.
5. Docs Keeper updates docs if behavior changed.

For tooling work, involve Tooling Keeper alongside Implementer or Tester.

## When To Split Work

Split work across roles when:

- the task touches more than one package
- behavior and docs need to change together
- editor tooling may be affected by syntax changes
- verification is non-trivial

Keep work unified when:

- the change is a small local fix
- only one package is affected
- test scope is obvious and narrow

## Success Criteria

A task is considered complete when:

- the target change is implemented
- the relevant verification has been run or explicitly deferred
- major risks are called out
- docs or examples are updated when needed
