package cmd

import (
	"fmt"

	"github.com/pocker/pocker/internal/config"
	"github.com/pocker/pocker/internal/image"
	"github.com/pocker/pocker/internal/skopeo"
	"github.com/spf13/cobra"
)

var saveOutput string
var saveTags []string

var saveCmd = &cobra.Command{
	Use:   "save IMAGE [PATH]",
	Short: "Save a remote image to a local docker-archive tar.gz",
	Long: `Save copies a remote image to docker-archive via skopeo.
The primary archive image name keeps the full reference (including ghcr.io/...).
Use --tag / -t to embed additional tags in the archive (primary tag is always kept).

  pocker save nginx:alpine
    → docker-archive:nginx-alpine.tar.gz:nginx:alpine

  pocker save ghcr.io/stefanprodan/podinfo:latest
    → docker-archive:…:ghcr.io/stefanprodan/podinfo:latest

  pocker save nginx:alpine /tmp/nginx.tar.gz
  pocker save nginx:alpine -o /tmp/nginx.tar.gz

  pocker save ghcr.io/username/image:tag -o out.tar.gz --tag username/images:tag
    → archive contains both tags`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := image.Parse(args[0])
		if err != nil {
			return err
		}
		out := saveOutput
		if len(args) == 2 {
			if out != "" && out != args[1] {
				return fmt.Errorf("conflicting output path: positional %q vs -o %q", args[1], out)
			}
			out = args[1]
		}
		primary := ref.NamedAs()
		extras, err := image.NormalizeExtraNames(primary, saveTags)
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
		src := ref.DockerSource()
		dest := ref.ArchiveDest(out)
		var opts []string
		for _, tag := range extras {
			opts = append(opts, "--additional-tag", tag)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Saving %s → %s\n", src, dest)
		if len(extras) > 0 {
			fmt.Fprintf(cmd.ErrOrStderr(), "Additional tags: %v\n", extras)
		}
		return runner.Copy(src, dest, opts...)
	},
}

func init() {
	saveCmd.Flags().StringVarP(&saveOutput, "output", "o", "", "output archive path (default: derived from image name)")
	saveCmd.Flags().StringArrayVarP(&saveTags, "tag", "t", nil, "additional archive image tag (repeatable; primary tag is always kept)")
}
