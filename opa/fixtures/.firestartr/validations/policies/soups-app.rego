package main

deny contains msg if {
    not startswith(input.name, data.data.prefixName)
    msg := sprintf("Field name must start with %v, but got: %v", [data.data.prefixName, input.name])
}

