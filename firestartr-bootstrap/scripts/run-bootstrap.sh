#!/bin/bash
dagger --bootstrap-file="./fixtures/Bootstrapfile.yaml" --credentials-secret="file:./Credentials.yaml" call run-bootstrap --docker-socket=/var/run/docker.sock --kind-svc=tcp://127.0.0.1:3000 $1
