# Copyright IBM Corp. 2015, 2026
# SPDX-License-Identifier: MPL-2.0

# Duplicate resources
resource "aws_instance" "foo" {}
resource "aws_instance" "foo" {}
