#!/bin/bash
dagger --firestartr-image="ghcr.io/prefapp/gitops-k8s" --firestartr-image-tag="v1.33.0_slim" --credentials-secret="file:./Credentials.yaml" call render-with-firestartr-container --claims-dir="./fixtures/claims" --crs-dir="./fixtures/crs" $1
