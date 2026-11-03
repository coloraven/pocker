package cmd

import (
	"fmt"

	"github.com/pocker/pocker/internal/config"
	"github.com/pocker/pocker/internal/image"
	"github.com/pocker/pocker/internal/manifest"
	"github.com/pocker/pocker/internal/skopeo"
	"github.com/spf13/cobra"
)

var pullTags []string
var pullFile string

var pullCmd = &cobra.Command{
	Use:   "pull [IMAGE | -f FILE]",
	Short: "Pull an image from a registry into the local Docker daemon",
	Long: `Pull copies a remote image into docker-daemon via skopeo.
The primary local tag matches the CLI reference (ghcr.io/... is preserved).
Use --tag / -t to add more local tags without dropping the primary one.
Additional tags are checked against the local daemon before pulling; if any
already exist, pull is aborted to avoid a post-download tagging failure.

Use --file / -f to read image references from a Dockerfile (FROM lines) or a
docker-compose.yml/.yaml (image: keys) and pull each of them. Build stage
names, build args and ${VAR} placeholders are skipped.

  pocker pull ubuntu:latest
    → skopeo copy docker://docker.io/library/ubuntu:latest docker-daemon:ubuntu:latest

  pocker pull ghcr.io/username/image:tag
    → docker-daemon:ghcr.io/username/image:tag

  pocker pull ghcr.io/username/image:tag --tag username/images:tag
    → primary: ghcr.io/username/image:tag
    → also:    username/images:tag

  pocker pull -f Dockerfile
  pocker pull -f docker-compose.yml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pullFile != "" && len(args) > 0 {
			return fmt.Errorf("cannot combine an image argument with --file %q", pullFile)
		}
		if pullFile == "" && len(args) == 0 {
			return fmt.Errorf("requires an IMAGE argument or --file FILE")
		}

		var refs []image.Ref
		if pullFile != "" {
			names, err := manifest.Images(pullFile)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return fmt.Errorf("no image references found in %q", pullFile)
			}
			for _, name := range names {
				ref, err := image.Parse(name)
				if err != nil {
					return fmt.Errorf("invalid image %q in %s: %w", name, pullFile, err)
				}
				refs = append(refs, ref)
			}
		} else {
			ref, err := image.Parse(args[0])
			if err != nil {
				return err
			}
			refs = append(refs, ref)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		runner, err := skopeo.New(cfg)
		if err != nil {
			return err
		}

		for i, ref := range refs {
			if len(refs) > 1 {
				fmt.Fprintf(cmd.ErrOrStderr(), "[%d/%d] %s\n", i+1, len(refs), ref.NamedAs())
			}
			if err := pullOne(cmd, runner, ref); err != nil {
				return err
			}
		}
		return nil
	},
}

func pullOne(cmd *cobra.Command, runner *skopeo.Runner, ref image.Ref) error {
	primary := ref.NamedAs()
	extras, err := image.NormalizeExtraNames(primary, pullTags)
	if err != nil {
		return err
	}
	if len(extras) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Checking additional tags against local daemon...\n")
		if err := runner.CheckExtraTagsFree(extras); err != nil {
			return err
		}
	}
	src := ref.DockerSource()
	dest := ref.DaemonDest()
	fmt.Fprintf(cmd.ErrOrStderr(), "Pulling %s → %s\n", src, dest)
	if err := runner.Copy(src, dest); err != nil {
		return err
	}
	for _, tag := range extras {
		fmt.Fprintf(cmd.ErrOrStderr(), "Tagging %s → %s\n", primary, tag)
		if err := runner.TagDaemon(primary, tag); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	pullCmd.Flags().StringArrayVarP(&pullTags, "tag", "t", nil, "additional local image tag (repeatable; primary tag is always kept)")
	pullCmd.Flags().StringVarP(&pullFile, "file", "f", "", "read images from a Dockerfile or docker-compose.yml/.yaml and pull each")
}
