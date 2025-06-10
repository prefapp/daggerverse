#!/bin/bash
dagger --bootstrap-file="./fixtures/Bootstrapfile.yaml" --credentials-secret="file:./Credentials.yaml" call github-repository-exists --repo="fake" --gh-token="env:GHTOKEN"
