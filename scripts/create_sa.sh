#!/bin/sh

gcloud iam service-accounts create "admin-sa" \
  --project "${PROJECT_ID}"