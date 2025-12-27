package kube

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Kubectl(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("kubectl %v failed: %s", args, msg)
	}
	return stdout.String(), nil
}

func ResolveServicePort(ctx context.Context, ns, svc, portName string, port int) (int, error) {
	if port != 0 {
		return port, nil
	}

	if strings.TrimSpace(portName) != "" {
		jsonpath := fmt.Sprintf("{.spec.ports[?(@.name=='%s')].port}", portName)
		out, err := Kubectl(ctx, "-n", ns, "get", "svc", svc, "-o", "jsonpath="+jsonpath)
		if err != nil {
			return 0, err
		}
		out = strings.TrimSpace(out)
		if out == "" {
			return 0, fmt.Errorf("service %s/%s has no port named '%s'", ns, svc, portName)
		}
		fields := strings.Fields(out)
		var p int
		if _, err := fmt.Sscanf(fields[0], "%d", &p); err != nil {
			return 0, fmt.Errorf("failed to parse port from %q: %w", out, err)
		}
		return p, nil
	}

	out, err := Kubectl(ctx, "-n", ns, "get", "svc", svc, "-o", "jsonpath={.spec.ports[0].port}")
	if err != nil {
		return 0, err
	}
	out = strings.TrimSpace(out)
	var p int
	if _, err := fmt.Sscanf(out, "%d", &p); err != nil {
		return 0, fmt.Errorf("failed to parse port from %q: %w", out, err)
	}
	return p, nil
}
