#!/bin/sh

gcloud iam workload-identity-pools describe "admin-pool" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --format="value(name)"