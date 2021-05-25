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

### Added

- `-v` command line tag that prints the version number and exits.

## v0.3.0 - 2021-05-17

### Added

- A new template context method `$.RequireArg` which returns the value of a
  template argument if it was provided, and fails the generation if the argument
  was not provided.
- A new template context method `$.DefaultArg` which returns the value of a
  template argument if it was provided, and a default value (second parameter
  passed to `DefaultArg`) if the template argument was not provided.

### Fixed

- The template method `typeName` now correctly returns the name of pointer types
  (without the asterisk).
- The template method `structField` can now return un-exported fields in
  addition to exported fields.

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
