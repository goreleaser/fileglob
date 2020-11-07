# fileglob

A file globbling library.

## What

`fileglob` is a glob library that uses [gobwas/glob](https://github.com/gobwas/glob) underneath
and returns only matching files.

## Why

[gobwas/glob](https://github.com/gobwas/glob) is very well implemented: it has
a lexer, compiler, and all that, which seems like a better approach than most
libraries do: regex.

It doesn't have a `Walk` method though, and we needed it
[in a couple of places](https://github.com/goreleaser/nfpm/issues/232).
So we decided to implement it ourselves, a little bit based on how
[mattn/go-zglob](http://github.com/mattn/go-zglob) works.
