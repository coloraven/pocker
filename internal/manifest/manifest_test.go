package manifest_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pocker/pocker/internal/manifest"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDockerfileImages(t *testing.T) {
	content := `# syntax=docker/dockerfile:1
ARG BASE=ubuntu:22.04
FROM ${BASE} AS builder
RUN echo hi
FROM golang:1.22-alpine AS build
FROM build
FROM --platform=$BUILDPLATFORM node:20-slim
from ghcr.io/owner/app:v1
`
	path := writeTemp(t, "Dockerfile", content)
	got, err := manifest.Images(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"golang:1.22-alpine", "node:20-slim", "ghcr.io/owner/app:v1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestComposeImages(t *testing.T) {
	content := `services:
  web:
    image: nginx:1.25-alpine
  db:
    image: "postgres:16"   # pinned
  cache:
    image: 'redis:7'
  app:
    build: .
  dyn:
    image: ${APP_IMAGE}
  empty:
    image:
`
	path := writeTemp(t, "docker-compose.yml", content)
	got, err := manifest.Images(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"nginx:1.25-alpine", "postgres:16", "redis:7"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestDetectUnknown(t *testing.T) {
	path := writeTemp(t, "notes.txt", "FROM ubuntu\n")
	if _, err := manifest.Images(path); err == nil {
		t.Fatal("expected error for unknown manifest type")
	}
}
