#!/bin/bash
dagger --bootstrap-file="./fixtures/Bootstrapfile.yaml" --credentials-secret="file:./Credentials.yaml" call clone-repo --repo="claims" --gh-token="env:GH_TOKEN" $1
