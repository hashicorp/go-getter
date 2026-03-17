## [2.0.0] - 2026-03-10

### ⚠️ BREAKING CHANGES

This is a **major version release** with **non-backward compatible changes**. Please review the migration guide before upgrading.

#### Removed Cryptographic Algorithms

- **Removed:** MD5 checksum support - No longer supported due to cryptographic weakness
- **Removed:** SHA1 checksum support - No longer supported due to cryptographic weakness
- **Supported:** SHA256 and SHA512 checksums only

Any URLs using `?checksum=md5:...` or `?checksum=sha1:...` parameters will now return an error.

#### Migration Required

Users must migrate all checksum references to use SHA256 or SHA512. See [MIGRATION.md](./MIGRATION.md) for detailed instructions.

#### Who Should Upgrade

- ✅ Upgrade to v2.0.0 if you can migrate all checksums to SHA256/SHA512
- ⏸️ Stay on v1.x if you require MD5/SHA1 support

Version 1.x will remain available for users who need MD5/SHA1 checksums temporarily.

### Security

- Enhanced security by removing weak cryptographic algorithms (MD5, SHA1)
- All checksum operations now use only cryptographically strong algorithms (SHA256, SHA512)

### Changed

- Checksum validation now exclusively supports SHA256 and SHA512 algorithms
- Error messages now guide users to use SHA256 or SHA512 for checksums
- Checksum guessing by length now only recognizes SHA256 (32 bytes) and SHA512 (64 bytes) lengths

### Removed

- Removed MD5 hash algorithm support (`crypto/md5` import)
- Removed SHA1 hash algorithm support (`crypto/sha1` import)
- Removed MD5-related test utilities and fixtures

### Fixed

- Improved security posture by eliminating weak hash algorithms

### Implementation Status
---

## [1.x] - Maintenance Branch

For users who require MD5 or SHA1 support, the v1.x branch will continue to receive critical security updates and bug fixes.

To use v1.x:
```bash
go get github.com/hashicorp/go-getter@v1
```
