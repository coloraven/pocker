package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pocker/pocker/internal/config"
	"github.com/spf13/cobra"
)

var configUnset bool

var configCmd = &cobra.Command{
	Use:   "config [KEY] [VALUE]",
	Short: "Get or set persistent pocker configuration",
	Long: `Manage persistent settings (proxies, skopeo path).

  pocker config                         # list all
  pocker config http_proxy              # get one key
  pocker config http_proxy http://...   # set one key
  pocker config all_proxy socks5://127.0.0.1:1080
  pocker config --unset http_proxy      # clear one key

Keys:
  http_proxy    HTTP(S) or SOCKS proxy (http://, https://, socks5://, socks5h://)
  https_proxy   HTTPS proxy (same schemes)
  all_proxy     SOCKS/all-traffic proxy (preferred for socks5:// / socks5h://)
  no_proxy      NO_PROXY list
  skopeo_path   Absolute path to skopeo binary

SOCKS notes:
  Set all_proxy to socks5://host:port (or socks5h:// for remote DNS).
  If http_proxy/https_proxy is a socks URL and all_proxy is empty,
  pocker also exports ALL_PROXY for skopeo.`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		path, err := config.Path()
		if err != nil {
			return err
		}

		if configUnset {
			if len(args) != 1 {
				return fmt.Errorf("--unset requires exactly one KEY")
			}
			if err := cfg.Unset(args[0]); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Printf("Unset %s\n", strings.ToLower(args[0]))
			fmt.Printf("Config: %s\n", path)
			return nil
		}

		switch len(args) {
		case 0:
			m := cfg.Map()
			if len(m) == 0 {
				fmt.Println("(no settings)")
				fmt.Printf("Config: %s\n", path)
				return nil
			}
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%s=%s\n", k, m[k])
			}
			fmt.Printf("Config: %s\n", path)
			return nil
		case 1:
			val, err := cfg.Get(args[0])
			if err != nil {
				return err
			}
			if val == "" {
				fmt.Printf("%s=(unset)\n", strings.ToLower(args[0]))
			} else {
				fmt.Printf("%s=%s\n", strings.ToLower(args[0]), val)
			}
			return nil
		default:
			if err := cfg.Set(args[0], args[1]); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Printf("%s=%s\n", strings.ToLower(args[0]), args[1])
			fmt.Printf("Saved: %s\n", path)
			return nil
		}
	},
}

func init() {
	configCmd.Flags().BoolVar(&configUnset, "unset", false, "unset a config key")
}
