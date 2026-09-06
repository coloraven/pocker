package image

import (
	"fmt"
	"strings"
)

// Ref is a parsed container image reference.
type Ref struct {
	// Raw is the user-provided reference (trimmed).
	Raw string
	// Registry is the registry host (e.g. docker.io, ghcr.io).
	Registry string
	// Repository is the path without registry and tag (e.g. library/ubuntu, username/image).
	Repository string
	// Tag is the image tag (default latest).
	Tag string
}

// Parse turns a Docker-style image name into a Ref.
//
// Examples:
//
//	ubuntu:latest              → docker.io/library/ubuntu:latest
//	nginx                      → docker.io/library/nginx:latest
//	user/app:1.0               → docker.io/user/app:1.0
//	ghcr.io/user/app:tag       → ghcr.io/user/app:tag
func Parse(raw string) (Ref, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Ref{}, fmt.Errorf("image reference is empty")
	}
	if strings.Contains(raw, "://") {
		return Ref{}, fmt.Errorf("unsupported transport in reference %q; use name[:tag]", raw)
	}

	name, tag := splitNameTag(raw)
	if name == "" {
		return Ref{}, fmt.Errorf("invalid image reference %q", raw)
	}

	registry, repository := splitRegistryRepo(name)
	if registry == "" {
		registry = "docker.io"
	}
	if registry == "docker.io" && !strings.Contains(repository, "/") {
		repository = "library/" + repository
	}

	if tag == "" {
		tag = "latest"
	}

	return Ref{
		Raw:        raw,
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
	}, nil
}

// DockerSource is the skopeo docker:// source URI.
func (r Ref) DockerSource() string {
	return fmt.Sprintf("docker://%s/%s:%s", r.Registry, r.Repository, r.Tag)
}

// NamedAs is the local image name: exactly what pull/save should tag as.
// Keeps registry prefixes (e.g. ghcr.io/...) so compose files can distinguish registries.
//
//	ubuntu:latest                 → ubuntu:latest
//	ghcr.io/username/image:tag    → ghcr.io/username/image:tag
func (r Ref) NamedAs() string {
	if r.Registry == "docker.io" {
		repo := r.Repository
		if strings.HasPrefix(repo, "library/") {
			repo = strings.TrimPrefix(repo, "library/")
		}
		return repo + ":" + r.Tag
	}
	return r.Registry + "/" + r.Repository + ":" + r.Tag
}

// DaemonDest is the skopeo docker-daemon: destination.
func (r Ref) DaemonDest() string {
	return "docker-daemon:" + r.NamedAs()
}

// LocalName is an alias of NamedAs (local docker tag).
func (r Ref) LocalName() string {
	return r.NamedAs()
}

// ArchiveFileName builds a default docker-archive filename from NamedAs.
//
//	nginx:alpine                       → nginx-alpine.tar.gz
//	ghcr.io/stefanprodan/podinfo:latest → ghcr.io-stefanprodan-podinfo-latest.tar.gz
func (r Ref) ArchiveFileName() string {
	safe := strings.NewReplacer("/", "-", ":", "-").Replace(r.NamedAs())
	return safe + ".tar.gz"
}

// ArchiveDest is the skopeo docker-archive: destination with optional override path.
// The embedded image name is NamedAs() so tags (including ghcr.io/...) are preserved.
func (r Ref) ArchiveDest(outPath string) string {
	if outPath == "" {
		outPath = r.ArchiveFileName()
	}
	return fmt.Sprintf("docker-archive:%s:%s", outPath, r.NamedAs())
}

func splitNameTag(ref string) (name, tag string) {
	if i := strings.LastIndex(ref, "@"); i >= 0 {
		return ref[:i], ref[i+1:]
	}
	i := strings.LastIndex(ref, ":")
	if i < 0 {
		return ref, ""
	}
	after := ref[i+1:]
	if strings.Contains(after, "/") {
		return ref, ""
	}
	return ref[:i], after
}

func splitRegistryRepo(name string) (registry, repository string) {
	slash := strings.Index(name, "/")
	if slash < 0 {
		return "", name
	}
	first := name[:slash]
	if isRegistryHost(first) {
		return first, name[slash+1:]
	}
	return "", name
}

func isRegistryHost(host string) bool {
	if host == "localhost" || strings.HasPrefix(host, "localhost:") {
		return true
	}
	if strings.Contains(host, ".") || strings.Contains(host, ":") {
		return true
	}
	return false
}
