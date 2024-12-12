dagger --deps-file=./fixtures/values-repo-dir/.github/hydrate_deps.yaml --values-dir=./fixtures/values-repo-dir/ --wet-repo-dir=./fixtures/wet-repo-dir/ call render-app --env=dev --tenant=test-tenant --app=sample-app --cluster=cluster-name --new-images-matrix='{ "images": [{ "service_name_list": ["micro-a", "micro-b"], "image": "test-image" }] }' stdout
