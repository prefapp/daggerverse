#!/bin/bash
dagger call render-provider-configs --templ=templates/providerconfigs.tmpl --creds=file:fixtures/Credsfile.yaml
