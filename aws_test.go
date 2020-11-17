package awsdd_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	awsdd "github.com/dharnitski/opentracing-aws-dd"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

func TestAWS(t *testing.T) {
	cfg := aws.NewConfig().
		WithRegion("us-west-2").
		WithDisableSSL(true).
		WithCredentials(credentials.AnonymousCredentials)

	session := awsdd.WrapSession(session.Must(session.NewSession(cfg)))

	t.Run("s3", func(t *testing.T) {
		mt := mocktracer.New()
		opentracing.SetGlobalTracer(mt)
		defer mt.Reset()

		root, ctx := opentracing.StartSpanFromContext(context.Background(), "test")
		s3api := s3.New(session)
		_, err := s3api.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: aws.String("BUCKET"),
		})
		assert.Error(t, err)
		root.Finish()

		spans := mt.FinishedSpans()
		require.Len(t, spans, 2)
		s := spans[0]
		s2 := spans[1]
		assert.Equal(t, s.SpanContext.TraceID, s2.SpanContext.TraceID)

		assert.Equal(t, "s3.command", s.OperationName)
		assert.Contains(t, s.Tag("aws.agent"), "aws-sdk-go")
		assert.Equal(t, "CreateBucket", s.Tag("aws.operation"))
		assert.Equal(t, "us-west-2", s.Tag("aws.region"))
		assert.Equal(t, "s3.CreateBucket", s.Tag(ext.ResourceName))
		assert.Equal(t, "aws.s3", s.Tag(ext.ServiceName))
		assert.Equal(t, "403", s.Tag(ext.HTTPCode))
		assert.Equal(t, "PUT", s.Tag(ext.HTTPMethod))
		assert.Equal(t, "http://s3.us-west-2.amazonaws.com/BUCKET", s.Tag(ext.HTTPURL))
	})

	t.Run("ec2", func(t *testing.T) {
		mt := mocktracer.New()
		opentracing.SetGlobalTracer(mt)
		defer mt.Reset()

		root, ctx := opentracing.StartSpanFromContext(context.Background(), "test")
		ec2api := ec2.New(session)
		_, err := ec2api.DescribeInstancesWithContext(ctx, &ec2.DescribeInstancesInput{})
		assert.Error(t, err)
		root.Finish()

		spans := mt.FinishedSpans()
		require.Len(t, spans, 2)
		s := spans[0]
		s2 := spans[1]

		assert.Equal(t, s.SpanContext.TraceID, s2.SpanContext.TraceID)
		assert.Equal(t, "ec2.command", s.OperationName)
		assert.Contains(t, s.Tag("aws.agent"), "aws-sdk-go")
		assert.Equal(t, "DescribeInstances", s.Tag("aws.operation"))
		assert.Equal(t, "us-west-2", s.Tag("aws.region"))
		assert.Equal(t, "ec2.DescribeInstances", s.Tag(ext.ResourceName))
		assert.Equal(t, "aws.ec2", s.Tag(ext.ServiceName))
		assert.Equal(t, "400", s.Tag(ext.HTTPCode))
		assert.Equal(t, "POST", s.Tag(ext.HTTPMethod))
		assert.Equal(t, "http://ec2.us-west-2.amazonaws.com/", s.Tag(ext.HTTPURL))
	})
}
