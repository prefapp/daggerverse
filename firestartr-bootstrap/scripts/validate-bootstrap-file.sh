#!/bin/bash
dagger --bootstrap-file="./fixtures/Bootstrapfile.yaml" --credentials-secret="file:./Credentials.yaml" call validate-bootstrap-file --bootstrap-file="./fixtures/Bootstrapfile.yaml"
