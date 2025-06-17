# Changelog

All notable changes to this project since forking from `github.com/mmcloughlin/trunnel` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Binary Encoder Functionality** - Complete implementation of binary encoding to complement existing parsing
  - `MarshalBinary() ([]byte, error)` method implementing Go's `encoding.BinaryMarshaler` interface
  - Internal `encodeBinary() []byte` method for efficient nested struct encoding
  - `validate() error` method for comprehensive constraint validation before encoding
  - Support for all trunnel data types:
    - Integer types (u8, u16, u32, u64) with big-endian encoding
    - Character type (char)
    - Null-terminated strings (nulterm)
    - Fixed arrays `type[N]`
    - Variable arrays `type[field]` 
    - Leftover arrays `type[]`
    - Nested struct references
    - Union types with conditional encoding
  - Comprehensive constraint validation:
    - Integer range constraints `IN [min..max]`
    - Integer value constraints `IN [val1, val2, ...]`
    - Array length validation
    - Union tag validation
  - Full round-trip compatibility: parse → encode → parse
  - Efficient encoding without redundant validation in nested structures
  - Proper handling of naming conflicts (e.g., fields named 'bytes')

### Changed
- **Code Generation Enhancement** - Extended `gen/decl.go` to generate encoder methods alongside existing parsers
- **Examples Updated** - All example packages now include encoder functionality
- **Test Coverage** - Comprehensive test suites added covering all data types and edge cases

### Technical Details
- **Big-endian encoding** consistent with trunnel specification
- **Automatic import resolution** via `golang.org/x/tools/imports`
- **Standard interface compliance** with Go's `encoding.BinaryMarshaler`
- **Performance optimized** buffer building and validation
- **Complex protocol support** tested with SOCKS5 protocol implementation

## [1.0.0] - 2025-06-16

### Added
- **Project Migration** - Forked and migrated from unmaintained `github.com/mmcloughlin/trunnel`
  - Initialized new Go module with path `github.com/katzenpost/trunnel`
  - Updated all import statements across entire codebase
  - Maintained full backward compatibility with existing trunnel functionality

### Changed
- **Module Path** - Changed from `github.com/mmcloughlin/trunnel` to `github.com/katzenpost/trunnel`
- **CLI Compatibility** - Fixed compatibility issues with urfave/cli library
  - Removed problematic short flags
  - Updated command structure for modern CLI library version
- **Build Configuration Updates**:
  - Updated Makefile PKG variable and project references
  - Updated .travis.yml for new project path and coverage reporting
  - Updated README.md badges, installation instructions, and project links
- **Dependencies** - Cleaned up and updated Go module dependencies
- **Test Infrastructure** - Regenerated test golden files to match new module structure

### Fixed
- **CLI Tool** - Resolved command-line interface issues preventing proper operation
- **Import Paths** - Fixed all internal import references to use new module path
- **Build System** - Ensured all build and test processes work with new project structure

### Migration Notes
This migration enables continued development and maintenance of the trunnel project under the katzenpost organization after the original repository became unmaintained for 4+ years. All existing functionality is preserved while enabling future enhancements.

## Pre-Fork History

For changes prior to the fork, please refer to the original repository at `github.com/mmcloughlin/trunnel`.

---

## Legend

- **Added** for new features
- **Changed** for changes in existing functionality  
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** for vulnerability fixes
