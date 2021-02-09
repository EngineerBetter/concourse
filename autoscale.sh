#!/usr/bin/env bash

: ${CONCOURSE_PROMETHEUS_ENDPOINT?"Must specify CONCOURSE_PROMETHEUS_ENDPOINT"}
: ${TEAMS?"Must specify TEAMS"}

current_metric=$(prom2json $CONCOURSE_PROMETHEUS_ENDPOINT)
retried_metrics=$(echo $current_metric | jq -c '.[] | select(.name == "concourse_workers_retried_errors").metrics')
container_metrics=$(echo $current_metric | jq -c '[.[] | select(.name == "concourse_workers_containers").metrics[] | select(.labels.worker | test("concourse-check-worker-.*") | not)]')

for team in $TEAMS; do
  retried_value=$(echo $retried_metrics | jq -r ".[] | select(.labels.team == \"$team\").value")
  container_value=$(echo $container_metrics | jq -r ".[] | select(.labels.worker | test(\"concourse-worker-$team-.*\")).value")
  retried_state_file=${team}-retried.metric

  if [[ $retried_value -gt $(cat $retried_state_file) ]]; then
    echo "scaling out workers for team $team"
#    kubectl scale statefulset/concourse-worker-joebloggs --replicas 2
  fi
  echo -n $retried_value > $retried_state_file

  echo "Container value: $container_value"
  if [[ $container_value == 0 ]]; then
    echo "scaling in workers for team $team"
#    kubectl scale statefulset/concourse-worker-joebloggs --replicas 2
  fi
done
