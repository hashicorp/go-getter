# Copyright IBM Corp. 2015, 2025
# SPDX-License-Identifier: MPL-2.0

module "child" {
    source = "./child"
}

resource "aws_instance" "foo" {
    memory = "${module.child.memory}"
}
