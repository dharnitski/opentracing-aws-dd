package awsdd_test

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	awsdd "github.com/dharnitski/opentracing-aws-dd"
)

func ExampleWrapSession() {
	session, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	// session is instrumented with global tracer
	session = awsdd.WrapSession(session)
}

func ExampleOption() {
	session, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	session = awsdd.WrapSession(session, awsdd.WithServiceName("myservice"))
}
