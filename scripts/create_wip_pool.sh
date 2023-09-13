#!/bin/sh

gcloud iam workload-identity-pools create "admin-pool" \
  --project="${PROJECT_ID}" \
  --location="global" \
  --display-name="Admin pool"