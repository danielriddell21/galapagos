package cli

import (
	"context"
	"fmt"
	"net/http"
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
		Long:  "Serves the WebAssembly build of the windowed demo over HTTP. This works on every platform because rendering happens in the browser, not via native OpenGL.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !webui.Bundled() {
				return fmt.Errorf("this build does not bundle the browser demo; install a release binary, or run `just wasm` and rebuild")
			}
			url := "http://" + addr
			fmt.Printf("serving the Galapagos browser demo at %s\n", url)
			if openPage {
				openBrowser(url)
			}
			srv := &http.Server{
				Addr:              addr,
				Handler:           webui.Handler(),
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
