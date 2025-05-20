#!/bin/bash
dagger --bootstrap-file="./fixtures/Bootstrapfile.yaml" --credentials-secret="file:./Credentials.yaml" call generate-github-token
