[![Build Status](https://travis-ci.org/dharnitski/opentracing-aws-dd.svg?branch=master)](https://travis-ci.org/dharnitski/opentracing-aws-dd)

# OpenTracing for AWS SDK in Go with Datadog schematic

This package is functional equivalent of [gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws](https://github.com/DataDog/dd-trace-go/tree/v1/contrib/aws/aws-sdk-go/aws) using opentracing framework. Although it is configured with Datadog tags schema there is nothing that prevents using it with different provider.

See example in [example_test.go](example_test.go)