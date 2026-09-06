package cmd

import (
	"fmt"

	"github.com/pocker/pocker/internal/config"
	"github.com/pocker/pocker/internal/image"
	"github.com/pocker/pocker/internal/skopeo"
	"github.com/spf13/cobra"
)

var pullTags []string

var pullCmd = &cobra.Command{
	Use:   "pull IMAGE",
	Short: "Pull an image from a registry into the local Docker daemon",
	Long: `Pull copies a remote image into docker-daemon via skopeo.
The primary local tag matches the CLI reference (ghcr.io/... is preserved).
Use --tag / -t to add more local tags without dropping the primary one.
Additional tags are checked against the local daemon before pulling; if any
already exist, pull is aborted to avoid a post-download tagging failure.

  pocker pull ubuntu:latest
    → skopeo copy docker://docker.io/library/ubuntu:latest docker-daemon:ubuntu:latest

  pocker pull ghcr.io/username/image:tag
    → docker-daemon:ghcr.io/username/image:tag

  pocker pull ghcr.io/username/image:tag --tag username/images:tag
    → primary: ghcr.io/username/image:tag
    → also:    username/images:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := image.Parse(args[0])
		if err != nil {
			return err
		}
		primary := ref.NamedAs()
		extras, err := image.NormalizeExtraNames(primary, pullTags)
		if err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		runner, err := skopeo.New(cfg)
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
	},
}

func init() {
	pullCmd.Flags().StringArrayVarP(&pullTags, "tag", "t", nil, "additional local image tag (repeatable; primary tag is always kept)")
}
