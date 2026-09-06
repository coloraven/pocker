package skopeo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pocker/pocker/internal/config"
)

// Runner executes skopeo with pocker config applied.
type Runner struct {
	Binary string
	Extra  []string
}

// New creates a Runner using config skopeo_path or PATH lookup.
func New(cfg *config.Config) (*Runner, error) {
	bin := ""
	if cfg != nil {
		bin = strings.TrimSpace(cfg.SkopeoPath)
	}
	if bin == "" {
		bin = "skopeo"
	}
	resolved, err := exec.LookPath(bin)
	if err != nil {
		hint := "install skopeo and ensure it is on PATH"
		if runtime.GOOS == "windows" {
			hint = "install skopeo for Windows (native, not WSL) and ensure skopeo.exe is on PATH, or set: pocker config skopeo_path C:\\path\\to\\skopeo.exe"
		}
		return nil, fmt.Errorf("skopeo not found (%s): %w; %s", bin, err, hint)
	}
	var extra []string
	if cfg != nil {
		extra = cfg.ProxyEnv()
	}
	return &Runner{Binary: resolved, Extra: extra}, nil
}

// Copy runs: skopeo copy [opts...] <src> <dest>
func (r *Runner) Copy(src, dest string, opts ...string) error {
	cmdArgs := append([]string{"copy"}, opts...)
	cmdArgs = append(cmdArgs, src, dest)
	return r.Run(cmdArgs...)
}

// TagDaemon copies a local docker-daemon image to another daemon name (add tag, no re-download).
func (r *Runner) TagDaemon(fromName, toName string) error {
	return r.Copy("docker-daemon:"+fromName, "docker-daemon:"+toName)
}

// DaemonExists reports whether name already exists in the local docker daemon.
func (r *Runner) DaemonExists(name string) (bool, error) {
	cmd := exec.Command(r.Binary, "inspect", "docker-daemon:"+name)
	cmd.Env = append(os.Environ(), r.Extra...)
	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	msg := stderr.String()
	if isDaemonNotFound(msg) {
		return false, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false, fmt.Errorf("inspect local image %q: %s", name, strings.TrimSpace(msg))
	}
	return false, fmt.Errorf("inspect local image %q: %w", name, err)
}

// CheckExtraTagsFree fails if any additional tag already exists in the local daemon.
func (r *Runner) CheckExtraTagsFree(tags []string) error {
	for _, tag := range tags {
		exists, err := r.DaemonExists(tag)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("additional tag %q already exists in local docker daemon; refuse to pull to avoid tagging conflict (remove or retag it first)", tag)
		}
	}
	return nil
}

func isDaemonNotFound(msg string) bool {
	s := strings.ToLower(msg)
	return strings.Contains(s, "no such image") ||
		strings.Contains(s, "not found") ||
		strings.Contains(s, "does not exist") ||
		strings.Contains(s, "no image found")
}

// Run executes skopeo with the given args, streaming stdout/stderr.
func (r *Runner) Run(args ...string) error {
	cmd := exec.Command(r.Binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), r.Extra...)
	fmt.Fprintf(os.Stderr, "==> %s %s\n", r.Binary, strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("skopeo failed: %w", err)
	}
	return nil
}
