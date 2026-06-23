package blobpath

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

type BlobPath struct {
	owner   string
	repo    string
	refPath string
}

var ErrInvalidPath = errors.New("invalid path")
var ErrMissingBlobPart = errors.New("missing blob part")

func Parse(p string) (*BlobPath, error) {
	p = path.Clean(p)
	parts := strings.Split(p, "/")
	if p[0] == '/' {
		parts = parts[1:]
	}

	if len(parts) < 5 {
		return nil, ErrInvalidPath
	}

	if parts[2] != "blob" {
		return nil, ErrMissingBlobPart
	}

	return &BlobPath{
		owner:   parts[0],
		repo:    parts[1],
		refPath: path.Join(parts[3:]...),
	}, nil
}

func (p *BlobPath) RawPath() string {
	return path.Join(p.owner, p.repo, p.refPath)
}

var commitLike = regexp.MustCompile(`(?i)^([0-9a-f]{7}|[0-9a-f]{40}|[0-9a-f]{64})$`)

func (p *BlobPath) MediaCandidates() []*MediaCandidate {
	parts := strings.Split(p.refPath, "/")

	for i := 0; i < len(parts); i++ {
		parts[i] = url.PathEscape(parts[i])
	}

	o := url.PathEscape(p.owner)
	r := url.PathEscape(p.repo)

	if len(parts) == 2 && commitLike.MatchString(parts[0]) {
		return []*MediaCandidate{{
			Owner: o,
			Repo:  r,
			Ref:   parts[0],
			Path:  parts[1],
		}}
	}

	var candidates []*MediaCandidate

	for i := len(parts) - 1; i >= 0; i-- {
		candidates = append(candidates, &MediaCandidate{
			Owner: o,
			Repo:  r,
			Ref:   url.QueryEscape(path.Join(parts[:i]...)),
			Path:  path.Join(parts[i:]...),
		})
	}

	return candidates
}

func (p *BlobPath) String() string {
	return path.Join(p.owner, p.repo, "blob", p.refPath)
}

type MediaCandidate struct {
	Owner string
	Repo  string
	Ref   string
	Path  string
}

func (mc *MediaCandidate) String() string {
	return fmt.Sprintf(
		"/repos/%s/%s/contents/%s?ref=%s",
		mc.Owner,
		mc.Repo,
		mc.Path,
		mc.Ref,
	)
}
