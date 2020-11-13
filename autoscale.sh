#!/usr/bin/env bash

: ${CONCOURSE_PROMETHEUS_ENDPOINT?"Must specify CONCOURSE_PROMETHEUS_ENDPOINT"}

current_metric=$(prom2json $CONCOURSE_PROMETHEUS_ENDPOINT)
metrics=$(echo $current_metric | jq -c '.[] | select(.name == "concourse_workers_retried_errors").metrics[]')

for metric in $metrics; do
  team=$(echo $metric | jq -r '.labels.team')
  value=$(echo $metric | jq -r '.value')
  state_file=${team}.metric

  if [ $value != $(cat $state_file) ]; then
    echo "scaling workers for team $team"
    kubectl scale statefulset/concourse-worker-joebloggs --replicas 2
  fi
  echo -n $value > $state_file
done