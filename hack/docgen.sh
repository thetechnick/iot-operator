#!/usr/bin/env bash

set -euo pipefail

cat << 'EOF' > ./docs/api-reference/_index.md
---
title: API Reference
weight: 50
---

# IoT Operator API Reference

The IoT Operator APIs are an extension of the [Kubernetes API](https://kubernetes.io/docs/reference/using-api/api-overview/) using `CustomResourceDefinitions`.

EOF

# IoT API Group
# --------------
cat << 'EOF' >> ./docs/api-reference/_index.md
## `iot.thetechnick.ninja`

The `iot.thetechnick.ninja` API group in contains all IoT related API objects.

EOF
find ./apis/iot/v1alpha1 -name '*types.go'  | xargs ./bin/docgen >> ./docs/api-reference/_index.md
