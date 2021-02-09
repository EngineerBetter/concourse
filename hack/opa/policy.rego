package concourse

default decision = {"allowed": true}

# uncomment to include deny rules
decision = {"allowed": false, "reasons": reasons} {
	count(deny) > 0
	reasons := deny
}

deny[msg] {
	input.action == "SelectWorker"
	input.data.container_spec.Type != "check"
	regex.match("concourse-check-worker-.*", input.data.selected_worker.name)
	msg := sprintf("only check containers can run on %s", [input.data.selected_worker.name])
}

deny[msg] {
	input.action == "SelectWorker"
	input.data.container_spec.Type == "check"
	not regex.match("concourse-check-worker-.*", input.data.selected_worker.name)
	msg := sprintf("check containers cannot run on %s", [input.data.selected_worker.name])
}

deny["cannot use docker-image types"] {
	input.action == "UseImage"
	input.data.image_type == "docker-image"
}

deny["cannot run privileged tasks"] {
	input.action == "SaveConfig"
	input.data.jobs[_].plan[_].privileged
}

deny["cannot use privileged resource types"] {
	input.action == "SaveConfig"
	input.data.resource_types[_].privileged
}
