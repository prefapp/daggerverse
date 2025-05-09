#!/bin/bash
dagger --firestartr-image="ghcr.io/prefapp/gitops-k8s" --bootstrap-file="./fixtures/Bootstrapfile.yaml" --firestartr-image-tag="v1.33.0_slim" --credentials-secret="file:./Credentials.yaml" call render-bootstrap-file --templ ./templates/components.tmpl
