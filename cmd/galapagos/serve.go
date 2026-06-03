package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/danielriddell21/galapagos/internal/webui"
	"github.com/spf13/cobra"
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
			return http.ListenAndServe(addr, webui.Handler())
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "localhost:8080", "address to serve on")
	cmd.Flags().BoolVar(&openPage, "open", true, "open the demo in a browser")
	rootCmd.AddCommand(cmd)
}

// openBrowser best-effort opens url in the default browser; failures are
// ignored since the URL is also printed.
func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	_ = exec.Command(cmd, append(args, url)...).Start()
}
