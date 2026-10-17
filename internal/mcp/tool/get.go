package tool

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/berquerant/git-iter-go/internal/config"
	"github.com/berquerant/git-iter-go/internal/git"
	"github.com/berquerant/git-iter-go/internal/output"
	"github.com/berquerant/git-iter-go/internal/repo"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetToolInput defines parameters for git_iter_get tool.
type GetToolInput struct {
	URI string `json:"uri,omitempty" jsonschema:"optional repository URI to resolve or clone into repos_root; if empty, lists discovered repositories"`
	RepoDiscoveryInput
	Timeout string `json:"timeout,omitempty" jsonschema:"per-process execution timeout for git clone (e.g. 5s, 1m)"`
}

// GetToolOutput defines output for git_iter_get tool.
type GetToolOutput struct {
	Repositories []output.RepoResult `json:"repositories" jsonschema:"resolved or discovered repositories"`
}

// RegisterGet registers the git_iter_get tool on the given MCP server.
func RegisterGet(server *mcpsdk.Server, cfg *config.Config) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "git_iter_get",
		Description: "Resolve or clone a git repository by URI under the root directory, or list discovered repositories if URI is omitted",
	}, func(
		ctx context.Context,
		_ *mcpsdk.CallToolRequest,
		in GetToolInput,
	) (*mcpsdk.CallToolResult, GetToolOutput, error) {
		if in.URI == "" {
			repos, err := listRepositories(ctx, cfg, in.RepoDiscoveryInput)
			if err != nil {
				return nil, GetToolOutput{}, err
			}
			return nil, GetToolOutput{Repositories: repos}, nil
		}

		root := in.ReposRoot
		if root == "" && cfg != nil {
			root = cfg.ReposRoot
		}

		targetDir, err := git.ResolveCloneDir(root, in.URI)
		if err != nil {
			return nil, GetToolOutput{}, err
		}

		if !repo.IsGitRepo(targetDir) {
			timeout, err := resolveTimeout(in.Timeout, cfg)
			if err != nil {
				return nil, GetToolOutput{}, err
			}

			gitCmd := ""
			if cfg != nil {
				gitCmd = cfg.GitCommandOrFallback()
			}
			g := git.New(gitCmd, nil)

			runCtx := ctx
			var cancel context.CancelFunc
			if timeout > 0 {
				runCtx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			if err := g.Clone(runCtx, in.URI, targetDir, io.Discard); err != nil {
				return nil, GetToolOutput{}, fmt.Errorf("clone %s: %w", in.URI, err)
			}
		}

		absRoot, _ := filepath.Abs(root)
		return nil, GetToolOutput{
			Repositories: []output.RepoResult{
				{
					RepoAbsPath: targetDir,
					RepoRelPath: output.RelPath(absRoot, targetDir),
				},
			},
		}, nil
	})
}
