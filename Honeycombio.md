---
title: "Honeycomb.io (Tracing)"
date: 2018-09-13
draft: true
weight: 3
class: "resized-logo"
aliases: [/supported-exporters/go/honeycomb]
logo: /img/honeycomb-logo.jpg
---

- [Introduction](#introduction)
- [Creating the exporter](#creating-the-exporter)
- [Viewing your traces](#viewing-your-traces)
- [Project link](#project-link)

## Introduction

Honeycomb is a saas tool that provides real-time system debugging. Visualize individual traces to deeply understand request executeion. Break down, filter and aggregate trace data to uncover patterns, find outliers, and understand historical trends.

Honeycomb aggregates data at read time, so you don’t have to predict ahead of time which metrics matter. You can fluidly go between time-series graphs, traces, and raw rows to get answers, without having to switch context between tools.

## Creating the exporter

To create the exporter, we'll need to:

- Create an exporter in code
- Honeycomb write key, found on your Honeycomb Team Settings page. ([Sign up for free](https://ui.honeycomb.io/signup) if you haven’t already!)

{{<highlight go>}}
package main

import (
    "log"
    
    libhoney "github.com/honeycombio/libhoney-go"
    honeycomb "github.com/honeycombio/opencensus-exporter/honeycomb"
    "go.opencensus.io/trace"

)

func main() {
    libhoney.Init(libhoney.Config{
		WriteKey: "YOUR WRITE KEY HERE",
		Dataset:  "YOUR DATASET HERE",
	})
	defer libhoney.Close()

    trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	br := bufio.NewReader(os.Stdin)

	trace.RegisterExporter(new(honeycomb.Exporter))
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1.0)})
}
{{</highlight>}}

## Viewing your traces

Please visit [honeycomb.io](https://ui.honeycomb.io/) to view your traces

## Project link

You can find out more about the Honeycomb at [https://www.honeycomb.io/](https://www.honeycomb.io/)
