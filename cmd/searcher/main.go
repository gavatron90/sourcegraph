// search is a simple service which exposes an API to text search a repo at
// a specific commit. See the searcher package for more information.
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/searcher/search"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")
var cacheDir = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
var cacheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "0", "maximum size of the on disk cache in megabytes")

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	tracer.Init("searcher")

	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err == nil {
		log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))
	}

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
	}

	var cacheSizeBytes int64
	if i, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		log.Fatalf("invalid int %q for SEARCHER_CACHE_SIZE_MB: %s", cacheSizeMB, err)
	} else {
		cacheSizeBytes = i * 1000 * 1000
	}

	service := &search.Service{
		Store: &search.Store{
			FetchTar:          fetchTar,
			Path:              filepath.Join(cacheDir, "searcher-archives"),
			MaxCacheSizeBytes: cacheSizeBytes,

			// Allow roughly 10 fetches per gitserver
			MaxConcurrentFetchTar: 10 * len(gitserver.DefaultClient.Addrs),
		},
	}
	if lvl >= log15.LvlInfo {
		service.RequestLog = log.New(os.Stderr, "", 0)
	}
	service.Store.Start()
	handler := nethttp.Middleware(opentracing.GlobalTracer(), service)

	addr := ":3181"
	server := &http.Server{Addr: addr, Handler: handler}
	go shutdownOnSIGINT(server)

	log15.Info("searcher: listening", "addr", ":3181")
	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func shutdownOnSIGINT(s *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctx)
	if err != nil {
		log.Fatal("graceful server shutdown failed, will exit:", err)
	}
}

func fetchTar(ctx context.Context, repo string, commit api.CommitID) (r io.ReadCloser, err error) {
	// gitcmd.Repository.Archive returns a zip file read into
	// memory. However, we do not need to read into memory and we want a
	// tar, so we directly run the gitserver Command.
	span, ctx := opentracing.StartSpanFromContext(ctx, "OpenTar")
	ext.Component.Set(span, "git")
	span.SetTag("URL", repo)
	span.SetTag("Commit", commit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err)
		}
		span.Finish()
	}()

	if strings.HasPrefix(string(commit), "-") {
		return nil, badRequestError{("invalid git revision spec (begins with '-')")}
	}

	cmd := gitserver.DefaultClient.Command("git", "archive", "--format=tar", string(commit))
	cmd.Repo = &api.Repo{URI: repo}
	cmd.EnsureRevision = string(commit)
	r, err = gitserver.StdoutReader(ctx, cmd)
	if err != nil {
		if vcs.IsRepoNotExist(err) || err == vcs.ErrRevisionNotFound {
			err = badRequestError{err.Error()}
		}
		return nil, err
	}
	return r, nil
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }
