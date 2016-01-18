package goapi;

// Error responders decide what kind of response to give when some error arise.
// The returned object will later be serialized (see ApiMethodfactory.errorResponder).
// By default, a simple struct response is given.
// If you use the default error responder with a custom serialiszer, it should be able to handle GeneralStatus.
// These responses are sent back to the user, so make sure to only send safe data.

// |err| may be nil.
type ErrorResponder func(err error, httpStatus int) interface{}

type GeneralStatus struct {
   Success bool
   Code int
}

func GeneralErrorResponder(err error, httpStatus int) interface{} {
   return GeneralStatus{false, httpStatus};
}
