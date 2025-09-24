package redirect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	clog "github.com/jasonhancock/cobra-logger"
	"github.com/jasonhancock/cobraflags/flags"
	"github.com/jasonhancock/cobraflags/root"
	ghttp "github.com/jasonhancock/go-http"
	"github.com/jasonhancock/go-logger"
	"github.com/spf13/cobra"
)

type options struct {
	HTTPAddr string
	DestAddr string
	HTTPCode int

	flags.FlagSet
}

func NewCmd(r *root.Command) *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:          "redirect",
		Short:        "Starts a redirect server.",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Check(); err != nil {
				return err
			}
			l := r.Logger(os.Stdout, root.WithName(clog.GetLoggerName(cmd)))
			return run(cmd.Context(), l, opts)
		},
	}

	opts.Add(
		cmd.Flags(),

		flags.New(
			&opts.HTTPAddr,
			"http-addr",
			"The interface and port to bind the http server to. If not set, an http server will not be started.",
			flags.Env("SERVER_ADDR"),
			flags.Required(),
		),

		flags.New(
			&opts.DestAddr,
			"dest-addr",
			"The destination URL to redirect to",
			flags.Env("DEST_ADDR"),
			flags.Required(),
		),

		flags.New(
			&opts.HTTPCode,
			"http-code",
			"The HTTP code to use when redirecting.",
			flags.Env("HTTP_CODE"),
			flags.Required(),
			flags.Default(http.StatusMovedPermanently),
		),
	)

	return cmd
}

func run(ctx context.Context, l *logger.L, opts options) error {
	dest, err := url.Parse(opts.DestAddr)
	if err != nil {
		return fmt.Errorf("parsing destination address: %w", err)
	}

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := *r.URL
		u.Host = dest.Host
		u.Scheme = dest.Scheme

		if dest.Path != "" {
			u.Path = path.Join(dest.Path, u.Path)
		}
		http.Redirect(w, r, u.String(), opts.HTTPCode)
	})

	var wg sync.WaitGroup
	ghttp.NewHTTPServer(
		ctx,
		l.New("http_server"),
		&wg,
		router,
		opts.HTTPAddr,
		ghttp.WithTimeeouts(10*time.Second),
	)

	wg.Wait()

	return nil
}
