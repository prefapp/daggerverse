package main

import (
	"context"
	"dagger/k-6/internal/dagger"
)

func (m *K6) ThrotthleCtr(
	ctx context.Context,
	// Container to throttle
	ctr *dagger.Container,
	// Bandwidth limit in megabytes per second
	networkThrottle int,
	// Network interface to throttle
	networkInterface string,
) *dagger.Container {
	container, _ := ctr.WithExec([]string{
		"apk",
		"update",
	}).WithExec([]string{
		"apk",
		"add",
		"iproute2-tc",
	}, dagger.ContainerWithExecOpts{
		InsecureRootCapabilities: true,
	}).WithExec([]string{
		"tc",
		"qdisc",
		"add",
		"dev", networkInterface,
		"handle", "ffff:", "ingress",
	}, dagger.ContainerWithExecOpts{
		InsecureRootCapabilities: true,
	}).Sync(ctx)

	return container
	/**

		WithExec([]string{
		"tc",
		"filter",
		"add",
		"dev", networkInterface,
		"protocol", "ip",
		"parent", "ffff:",
		"prio", "50",
		"u32",
		"match", "ip", "src", "0.0.0.0/0",
		"police", fmt.Sprintf("rate %dkbit burst 15k drop", networkThrottle),
		"flowid", ":1",
	}, dagger.ContainerWithExecOpts{
		InsecureRootCapabilities: true,
	}).WithExec([]string{

		"tc",
		"qdisc",
		"add",
		"dev", networkInterface,
		"root", "tbf",
		"rate", fmt.Sprintf("%dmbit", networkThrottle),
		"burst", "32kbit",
		"latency", "400ms",
	}, dagger.ContainerWithExecOpts{
		InsecureRootCapabilities: true,
	}).Sync(ctx)
	**/

}
