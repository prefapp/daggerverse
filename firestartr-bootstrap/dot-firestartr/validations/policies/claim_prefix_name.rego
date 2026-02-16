package main

deny contains msg if {
    not startswith(input.name, data.data.prefixName)
    msg := sprintf("Claim name must start with '%v', but got: '%v'", [data.data.prefixName, input.name])
}
