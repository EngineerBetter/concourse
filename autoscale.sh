#!/usr/bin/env bash

set -euo pipefail

: ${CONCOURSE_PROMETHEUS_ENDPOINT?"Must specify CONCOURSE_PROMETHEUS_ENDPOINT"}
: ${TEAMS?"Must specify TEAMS"}

current_metric=$(prom2json "${CONCOURSE_PROMETHEUS_ENDPOINT}")
retried_metrics=$(echo "${current_metric}" | jq -c '.[] | select(.name == "concourse_workers_retried_errors").metrics')
container_metrics=$(echo "${current_metric}" | jq -c '[.[] | select(.name == "concourse_workers_containers").metrics[] | select(.labels.worker | test("concourse-check-worker-.*") | not)]')

for team in $TEAMS; do
  retried_values=$(echo "${retried_metrics}" | jq -r ".[] | select(.labels.team == \"$team\") | \"\(.labels.error_type) \(.value)\"" | sort -k1)
  container_value=$(echo "${container_metrics}" | jq -r ".[] | select(.labels.worker | test(\"concourse-worker-$team-.*\")).value")
  retried_state_file=${team}-retried.metric

  if [[ ! -f "${retried_state_file}" ]]; then
    echo -n "${retried_values}" > "${retried_state_file}"
  fi

  if [[ "${retried_values}" != $(cat "${retried_state_file}") ]]; then
    echo "scaling out workers for team $team"
#    kubectl scale statefulset/concourse-worker-joebloggs --replicas 2
    echo -n "${retried_values}" > "${retried_state_file}"
    exit 0
  fi

  echo "Container value: $container_value"
  if [[ $container_value == 0 ]]; then
    echo "scaling in workers for team $team"
#    kubectl scale statefulset/concourse-worker-joebloggs --replicas 2
  exit 0
  fi
done
