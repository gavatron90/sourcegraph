	"github.com/sourcegraph/sourcegraph/pkg/vcs/git/gittest"
			repo: gittest.MakeGitRepository(t, gitCommands...),
							Author:    git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Author:    git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
							Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
							Author:    git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
							Author:    git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
							Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: gittest.MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				t.Errorf("%s: %+v: got %+v, want %+v", label, *opt, gittest.AsJSON(results), gittest.AsJSON(want))
			repo: gittest.MakeGitRepository(t, gitCommands...),
				t.Errorf("%s: %+v: got %+v, want %+v", label, *opt, gittest.AsJSON(results), gittest.AsJSON(want))