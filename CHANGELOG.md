
> **Note:** For previous release notes, see the [GitHub Releases page](https://github.com/hashicorp/go-getter/releases).

## Unreleased

### Improvements
- Improved error messages: better capitalization and error wrapping (#578)
- Enhanced test coverage for ambiguous git refs (#382) and checksum file handling (#250)

### Changes
- Updated build commands to specify the target directory for Go binaries (#594)
- Bumped Go modules dependencies including cloud.google.com/go/storage, aws-sdk-go-v2, and others (#593)
- Updated GitHub Actions versions

### Fixed
- Fixed HTTP file download being skipped when Content-Length is 0 (#538, #539)
- Fixed compilation errors in get_git_test.go (#579)

### Security
