package image_test

import (
	"testing"

	"github.com/pocker/pocker/internal/image"
)

func TestParseAndTransports(t *testing.T) {
	tests := []struct {
		raw     string
		source  string
		daemon  string
		archive string
		local   string
	}{
		{
			raw:     "ubuntu:latest",
			source:  "docker://docker.io/library/ubuntu:latest",
			daemon:  "docker-daemon:ubuntu:latest",
			archive: "docker-archive:ubuntu-latest.tar.gz:ubuntu:latest",
			local:   "ubuntu:latest",
		},
		{
			raw:     "nginx:alpine",
			source:  "docker://docker.io/library/nginx:alpine",
			daemon:  "docker-daemon:nginx:alpine",
			archive: "docker-archive:nginx-alpine.tar.gz:nginx:alpine",
			local:   "nginx:alpine",
		},
		{
			raw:     "ghcr.io/username/image:tag",
			source:  "docker://ghcr.io/username/image:tag",
			daemon:  "docker-daemon:ghcr.io/username/image:tag",
			archive: "docker-archive:ghcr.io-username-image-tag.tar.gz:ghcr.io/username/image:tag",
			local:   "ghcr.io/username/image:tag",
		},
		{
			raw:     "ghcr.io/stefanprodan/podinfo:latest",
			source:  "docker://ghcr.io/stefanprodan/podinfo:latest",
			daemon:  "docker-daemon:ghcr.io/stefanprodan/podinfo:latest",
			archive: "docker-archive:ghcr.io-stefanprodan-podinfo-latest.tar.gz:ghcr.io/stefanprodan/podinfo:latest",
			local:   "ghcr.io/stefanprodan/podinfo:latest",
		},
		{
			raw:     "user/app:1.0",
			source:  "docker://docker.io/user/app:1.0",
			daemon:  "docker-daemon:user/app:1.0",
			archive: "docker-archive:user-app-1.0.tar.gz:user/app:1.0",
			local:   "user/app:1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			ref, err := image.Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got := ref.DockerSource(); got != tt.source {
				t.Fatalf("DockerSource: got %q want %q", got, tt.source)
			}
			if got := ref.DaemonDest(); got != tt.daemon {
				t.Fatalf("DaemonDest: got %q want %q", got, tt.daemon)
			}
			if got := ref.ArchiveDest(""); got != tt.archive {
				t.Fatalf("ArchiveDest: got %q want %q", got, tt.archive)
			}
			if got := ref.LocalName(); got != tt.local {
				t.Fatalf("LocalName: got %q want %q", got, tt.local)
			}
			if got := ref.ArchiveDest("/tmp/out.tar.gz"); got != "docker-archive:/tmp/out.tar.gz:"+tt.local {
				t.Fatalf("ArchiveDest with path: got %q", got)
			}
		})
	}
}

func TestParseRegistryPort(t *testing.T) {
	ref, err := image.Parse("localhost:5000/my/app:v1")
	if err != nil {
		t.Fatal(err)
	}
	if ref.NamedAs() != "localhost:5000/my/app:v1" {
		t.Fatalf("NamedAs=%q", ref.NamedAs())
	}
	if ref.DaemonDest() != "docker-daemon:localhost:5000/my/app:v1" {
		t.Fatalf("daemon=%q", ref.DaemonDest())
	}
}
