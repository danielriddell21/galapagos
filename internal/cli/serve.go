package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/webui"
)

func init() {
	var (
		addr     string
		openPage bool
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the browser (WebAssembly) demo",
		Long: "Compiles the windowed demo to WebAssembly and serves it over HTTP. " +
			"This works on every platform because rendering happens in the browser, not via native OpenGL. " +
			"The demo is built on demand rather than shipped inside the binary, so this needs Go and a checkout of the repository.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// A temporary directory, because the demo is cheap to rebuild and
			// a copy left lying about is a copy that goes stale.
			dir, err := os.MkdirTemp("", "galapagos-web-")
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}
			defer func() { _ = os.RemoveAll(dir) }()

			fmt.Println("compiling the browser demo...")
			if err := webui.Build(cmd.Context(), dir); err != nil {
				return err
			}

			url := "http://" + addr
			fmt.Printf("serving the Galapagos browser demo at %s\n", url)
			if openPage {
				openBrowser(url)
			}
			srv := &http.Server{
				Addr:              addr,
				Handler:           webui.Handler(dir),
				ReadHeaderTimeout: 10 * time.Second,
			}
			if err := srv.ListenAndServe(); err != nil {
				return fmt.Errorf("serve: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "localhost:8080", "address to serve on")
	cmd.Flags().BoolVar(&openPage, "open", true, "open the demo in a browser")
	rootCmd.AddCommand(cmd)
}

func openBrowser(url string) {
	var cmd string
	args := make([]string, 0, 1)
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	_ = exec.CommandContext(context.Background(), cmd, append(args, url)...).Start()
}
