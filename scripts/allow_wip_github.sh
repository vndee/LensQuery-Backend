#!/bin/sh

export PROJECT_ID=lensquery
export REPO=vndee/LensQuery-Backend

gcloud iam service-accounts add-iam-policy-binding "lensquery@${PROJECT_ID}.iam.gserviceaccount.com" \
  --project="${PROJECT_ID}" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/${WORKLOAD_IDENTITY_POOL_ID}/attribute.repository/${REPO}"
