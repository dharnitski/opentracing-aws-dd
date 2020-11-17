package awsdd

import (
	"math"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	ddext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

const (
	tagAWSAgent     = "aws.agent"
	tagAWSOperation = "aws.operation"
	tagAWSRegion    = "aws.region"
)

type handlers struct {
	cfg *config
}

// WrapSession wraps a session.Session, causing requests and responses to be traced.
func WrapSession(s *session.Session, opts ...Option) *session.Session {
	cfg := new(config)
	defaults(cfg)
	for _, opt := range opts {
		opt(cfg)
	}
	h := &handlers{cfg: cfg}
	s = s.Copy()
	s.Handlers.Send.PushFrontNamed(request.NamedHandler{
		Name: "github.com/dharnitski/opentracing-aws-dd/handlers.Send",
		Fn:   h.Send,
	})
	s.Handlers.Complete.PushBackNamed(request.NamedHandler{
		Name: "github.com/dharnitski/opentracing-aws-dd/handlers.Complete",
		Fn:   h.Complete,
	})
	return s
}

func (h *handlers) Send(req *request.Request) {
	opts := []opentracing.StartSpanOption{
		opentracing.Tag{Key: ddext.SpanType, Value: ddext.SpanTypeHTTP},
		opentracing.Tag{Key: ddext.ServiceName, Value: h.serviceName(req)},
		opentracing.Tag{Key: ddext.ResourceName, Value: h.resourceName(req)},
		opentracing.Tag{Key: tagAWSAgent, Value: h.awsAgent(req)},
		opentracing.Tag{Key: tagAWSOperation, Value: h.awsOperation(req)},
		opentracing.Tag{Key: tagAWSRegion, Value: h.awsRegion(req)},
		opentracing.Tag{Key: ddext.HTTPMethod, Value: req.Operation.HTTPMethod},
		opentracing.Tag{Key: ddext.HTTPURL, Value: req.HTTPRequest.URL.String()},
	}
	if !math.IsNaN(h.cfg.analyticsRate) {
		opts = append(opts, opentracing.Tag{Key: ddext.EventSampleRate, Value: h.cfg.analyticsRate})
	}
	_, ctx := opentracing.StartSpanFromContext(req.Context(), h.operationName(req), opts...)
	req.SetContext(ctx)
}

func (h *handlers) Complete(req *request.Request) {
	span := opentracing.SpanFromContext(req.Context())
	if req.HTTPResponse != nil {
		span.SetTag(ddext.HTTPCode, strconv.Itoa(req.HTTPResponse.StatusCode))
	}
	if req.Error != nil {
		ext.LogError(span, req.Error)
	}
	span.Finish()
}

func (h *handlers) operationName(req *request.Request) string {
	return h.awsService(req) + ".command"
}

func (h *handlers) resourceName(req *request.Request) string {
	return h.awsService(req) + "." + req.Operation.Name
}

func (h *handlers) serviceName(req *request.Request) string {
	if h.cfg.serviceName != "" {
		return h.cfg.serviceName
	}
	return "aws." + h.awsService(req)
}

func (h *handlers) awsAgent(req *request.Request) string {
	if agent := req.HTTPRequest.Header.Get("User-Agent"); agent != "" {
		return agent
	}
	return "aws-sdk-go"
}

func (h *handlers) awsOperation(req *request.Request) string {
	return req.Operation.Name
}

func (h *handlers) awsRegion(req *request.Request) string {
	return req.ClientInfo.SigningRegion
}

func (h *handlers) awsService(req *request.Request) string {
	return req.ClientInfo.ServiceName
}
