# golambda 

[![Build](https://github.com/nhost/golambda/actions/workflows/build.yaml/badge.svg)](https://github.com/nhost/golambda/actions/workflows/build.yaml) &ensp; [![Release](https://github.com/nhost/golambda/actions/workflows/release.yaml/badge.svg?branch=main&event=release)](https://github.com/nhost/golambda/actions/workflows/release.yaml)

Simple golang utility to convert existing golang functions into lambda compatible ones.

We built this tool for our internal requirement of deploying Nhost functions using golang as their runtimes on AWS Lambda.

Checkout a simple example [here](/example).

## Install

Download the compiled binary from the [release page](https://github.com/nhost/golambda/releases).

Or download using go: `go get github.com/nhost/golambda`

## Usage

`golambda -source {golang_function_file}.go -output {output_zip_file}.zip`

The output file can directly be deployed on AWS Lambda.
