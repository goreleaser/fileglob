<p align="center">
  <img alt="GoReleaser Logo" src="https://avatars2.githubusercontent.com/u/24697112?v=3&s=200" height="140" />
  <h1 align="center">fileglob</h1>
  <p align="center">A file globbling library.</p>
  <p align="center">
    <a href="https://github.com/goreleaser/fileglob/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/goreleaser/fileglob.svg?style=for-the-badge"></a>
    <a href="/LICENSE.md"><img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge"></a>
    <a href="https://github.com/goreleaser/fileglob/actions?workflow=build"><img alt="GitHub Actions" src="https://img.shields.io/github/workflow/status/goreleaser/fileglob/build?style=for-the-badge"></a>
    <a href="https://codecov.io/gh/goreleaser/fileglob"><img alt="Codecov branch" src="https://img.shields.io/codecov/c/github/goreleaser/fileglob/master.svg?style=for-the-badge"></a>
    <a href="https://goreportcard.com/report/github.com/goreleaser/fileglob"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/goreleaser/fileglob?style=for-the-badge"></a>
    <a href="http://godoc.org/github.com/goreleaser/fileglob"><img alt="Go Doc" src="https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge"></a>
    <a href="https://github.com/goreleaser"><img alt="Powered By: GoReleaser" src="https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge"></a>
  </p>
</p>

## What

`fileglob` is a glob library that uses [gobwas/glob](https://github.com/gobwas/glob) underneath
and returns only matching files.

## Why

[gobwas/glob](https://github.com/gobwas/glob) is very well implemented: it has
a lexer, compiler, and all that, which seems like a better approach than most
libraries do: regex.

It doesn't have a `Walk` method though, and we needed it
[in a couple of places](https://github.com/goreleaser/fileglob/issues/232).
So we decided to implement it ourselves, a little bit based on how
[mattn/go-zglob](http://github.com/mattn/go-zglob) works.
