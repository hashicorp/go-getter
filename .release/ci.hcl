schema = "1"

project "go-getter" {
  team = "team-ip-compliance"

  slack {
    notification_channel = "C09KTF77K6X" // ensure this is a PUBLIC slack channel. If it's private, the promotion workflows will fail.
  }

  github {
    organization     = "hashicorp"
    repository       = "go-getter"
    release_branches = ["main", "release/**"]
  }
}

event "merge" {
}
event "build" {

  action "build" {
    organization = "hashicorp"
    repository   = "go-getter"
    workflow     = "build"
    depends      = null
    config       = ""
  }

  depends = ["merge"]
}
event "prepare" {

  action "prepare" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "prepare"
    depends      = ["build"]
    config       = ""
  }

  depends = ["build"]

  notification {
    on = "fail"
  }
}
event "trigger-staging" {
}
event "promote-staging" {

  action "promote-staging" {
    organization = "hashicorp"
    repository   = "go-getter"
    workflow     = "promote-staging"
    depends      = null
    config       = "oss-release-metadata.hcl"
  }

  depends = ["trigger-staging"]

  notification {
    on = "always"
  }

  promotion-events {

    pre-promotion {
      organization = "hashicorp"
      repository   = "go-getter"
      workflow     = "enos-run"
    }
  }
}
event "trigger-production" {
}
event "promote-production" {

  action "promote-production" {
    organization = "hashicorp"
    repository   = "crt-workflows-common"
    workflow     = "promote-production"
    depends      = null
    config       = ""
  }

  depends = ["trigger-production"]

  notification {
    on = "always"
  }

  promotion-events {
    bump-version-patch = true
    update-ironbank    = true
  }
}