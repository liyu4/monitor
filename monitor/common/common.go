package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

// GetPid Get single pid.
func GetPid(name string) (string, error) {
	ctx, cancel := context.WithCancel(context.TODO())

	cmd := exec.CommandContext(ctx, "pgrep", "-f", name)

	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(stdout)

	if err != nil {
		return "", err
	}

	cmd.Wait()
	cancel()
	select {
	case <-ctx.Done():
	}

	a := strings.Split(string(bytes), "\n")

	if len(a) >= 1 {
		return a[0], nil
	}

	return "", fmt.Errorf("GetPid error: %v", "get pid failed!")

}
