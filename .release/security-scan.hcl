# Copyright IBM Corp. 2015, 2026

binary {
	secrets      = true
	go_modules   = true
	osv          = true
	oss_index    = false
	nvd          = false
	go_stdlib    = true

  # Triage items that are _safe_ to ignore here. Note that this list should be
  # periodically cleaned up to remove items that are no longer found by the
  # scanner.
  triage {
    suppress {
      vulnerabilities = [
        // Ref: https://pkg.go.dev/vuln/GO-2026-5932
        //
        // Exists in the "golang.org/x/crypto/openpgp" package which is not used
        // by go-getter and is deprecated and unmaintained.
        "GO-2026-5932",
      ]
    }
  }
}
