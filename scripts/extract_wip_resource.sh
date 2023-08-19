#!/bin/sh

gcloud iam workload-identity-pools providers describe "github-provider" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --workload-identity-pool="lensquery-cd" \
  --format="value(name)"
