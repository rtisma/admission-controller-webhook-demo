#!/bin/bash
kube="kubectl -n webhook-demo"

${kube} delete svc/webhook-server
${kube} delete deploy/webhook-server
${kube} delete mutatingwebhookconfigurations/demo-webhook
${kube} delete secret webhook-server-tls
${kube} get pods -ojson | jq -r .items[].metadata.labels.runName | xargs ${kube} delete pods
