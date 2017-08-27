package goapi;

import (
   "fmt"
)

type ApiMethodFactory struct {
   contentType string
   log Logger
   serializer Serializer
   errorResponder ErrorResponder
   tokenValidator ValidateToken
}

func (factory *ApiMethodFactory) SetLogger(log Logger) {
   factory.log = log;
}

func (factory *ApiMethodFactory) SetSerializer(serializer Serializer) {
   factory.serializer = serializer;
}

func (factory *ApiMethodFactory) SetContentType(contentType string) {
   factory.contentType = contentType;
}

func (factory *ApiMethodFactory) SetGeneralErrorResponser(responder ErrorResponder) {
   factory.errorResponder = responder;
}

func (factory *ApiMethodFactory) SetTokenValidator(validator ValidateToken) {
   factory.tokenValidator = validator;
}

// Ensure that defaults are set if there are no user-supplied values.
func (factory *ApiMethodFactory) setDefaults() {
   if (factory.log == nil) {
      factory.log = ConsoleLogger{};
   }

   if (factory.serializer == nil) {
      factory.serializer = JSONSerializer;
   }

   if (factory.contentType == "") {
      factory.contentType = "application/json; charset=UTF-8";
   }

   if (factory.errorResponder == nil) {
      factory.errorResponder = GeneralErrorResponder;
   }
}

func (factory ApiMethodFactory) NewApiMethod(path string, handler interface{}, auth bool, params []ApiMethodParam) *ApiMethod {
   (&factory).setDefaults();

   // Ensure that there is a token validator if authentication is requested.
   if (auth && factory.tokenValidator == nil) {
      factory.log.Panic(fmt.Sprintf("API method for [%s] expects authentication, but no token authentication function has been set (see ApiMethodFactory.SetTokenValidator())", path));
   }

   var method ApiMethod = ApiMethod{
      path: path,
      handler: handler,
      auth: auth,
      params: params,
      log: factory.log,
      allowTokenParam: false,
      serializer: factory.serializer,
      contentType: factory.contentType,
      errorResponder: factory.errorResponder,
      tokenValidator: factory.tokenValidator,
   };

   method.validate();
   return &method;
}
