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

## v0.7.0 - 2022-05-31

### Added

- New template function, `catNoSpace`, that concatenates multiple strings
  together into one with no separators between them.  This complements the
  already included `cat` function from Sprig
  http://masterminds.github.io/sprig/strings.html which joins with a space in
  between each string.

## v0.6.0 - 2022-05-14

### Added

- New template function `isExported` that detects if a name is private or exported.

### Changed

- The template function `typeString` has been moved to a context method
  `$.TypeString` in order to prevent including the package in the type string if
  it is part of the current package.

### Fixed

- The current package is now always prevented from being imported into the
  generated file, even if it is passed to `AddImport`.
- Fixed parsing of the "codegen" tags to be compatible with go 1.17.

## v0.5.0 - 2022-05-05

### Added

- New template context method `$.AddImportType` which takes a type object
  instead of a string.
- New template method `typeString` which returns the full go representation of
  the type, including '*' for pointers, '[]' for slices, etc.

### Fixed

- `typeName` no longer panics for basic types that don't have a package: `int`,
  `string`, etc.

## v0.4.0 - 2021-05-25

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
