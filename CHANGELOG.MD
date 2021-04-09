# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Types of changes:

- `Security` in case of vulnerabilities.
- `Removed` for now removed features.
- `Deprecated` for soon-to-be removed features.
- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Fixed` for any bug fixes.
- `Internal` for changes that are worth noting, but don't have any external effects.

## Unreleased

## v0.2.0 - 2021-04-08

### Added

- A new template method `structFields` which returns an array of the fields
  (type `[]*types.Var`) declared on the passed struct.  Note that this doesn't
  include fields elevated from embedded types.

### Changed

- The template variable `.Struct` is now an instance of `*types.Named` instead
  of `*types.Struct`.  This is necessary to inspect the methods on the struct as
  they are tied to the named type, not the underlying struct.

## v0.1.0 - 2021-04-05

This is where we formally began tracking changes.
