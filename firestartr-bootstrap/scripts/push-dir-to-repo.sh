#!/bin/bash
dagger --bootstrap-file="./fixtures/Bootstrapfile.yaml" --credentials-secret="file:./Credentials.yaml" call push-dir-to-repo --dir="./fixtures/push_files_test" --repo-name="test-new-repo" --gh-token="env:GH_TOKEN"



