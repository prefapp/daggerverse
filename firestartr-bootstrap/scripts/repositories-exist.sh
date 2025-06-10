#!/bin/bash
dagger --bootstrap-file="./fixtures/Bootstrapfile.yaml" --credentials-secret="file:./Credentials.yaml" call validate-repositories-are-not-created-yet --gh-token="env:GHTOKEN"
