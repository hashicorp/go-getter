# Release process

This project is released via HashiCorp Common Release Tooling (CRT).

Follow internal documentation for full release details, but the high points are:

1. Make a PR to prepare the main branch.
   - `CHANGELOG.md` - update header, populate with changes as needed
   - `version/version.go` - change `"dev"` prerelease to `""`
1. Merging the PR starts a run of [`build.yml`][] in this repo
1. Completion of `build.yml` triggers a `prepare` run in the [CRT repo][]
   - Wait for `prepare` to complete
1. Run `bob trigger-promotion` to promote to staging
   - Wait for `promote-staging` in CRT repo to complete
1. Run `bob` again to promote to production
   - Wait for `promote-production` in CRT repo to complete
1. Make another PR to set up for next release
   - `CHANGELOG.md` - add `## Unreleased`
   - `version/VERSION` - bump version
   - `version/version.go` - bump version, add `"dev"` back to prerelease

[`build.yml`]: https://github.com/hashicorp/go-getter/actions/workflows/build.yml
[CRT repo]: https://github.com/hashicorp/crt-workflows-common/actions
