# goapi [![Build Status](https://travis-ci.org/eriq-augustine/goapi.svg?branch=master)](https://travis-ci.org/eriq-augustine/goapi)
A quick framework for making go-based HTTP APIs.

## Creating An API

An API is a collection of ApiMethods.

An ApiMethod is a description of a method that can be called on the API.
It includes a path, whether or nor to authenticate, handler, and parameters to the handler.

ApiMethods make it easy and safe to create API methods.
Parameters (existence, required, and type), http semantics, and authorization is all handled for you.

### API Parameters

An ApiMethod's parameters will be validated and passed into the ApiMethod's handler in the order that they were defined.
The name in the definition will be the name of the query http request parameter.
This value can come from query parameters or POST form values.
All types must match EXACTLY (eg integers must be 'int' not 'int64' or '*int').
Parameters may either be ints (API_PARAM_TYPE_INT), strings (API_PARAM_TYPE_STRING), or goapi.File (API_PARAM_TYPE_FILE).
Note that if you want to upload files you must use PUT or POST.
Base64 file data may be passed as a string instead.
Feel free to pass JSON in your string.
All parameters will be trimmed of whitespace before processing.

Note that ApiMethods are not allowed to have empty strings as parameters.
An empty string will be treated as a missing parameter.
In a similar vein, non-required ints are dangerous because 0 will be used as the empty value.
If you need to have a non-required int, consider typing it as a string and inspecting it manually.

Because refection in Go does not allow you to find the parameter's name, we must rely on order.

#### Handling Files

If you want to upload a file, you can either use a base64 encoded string or do a POST/PUT with Content-Type = multipart/form-data.
This section describes details about the later.
On the server side, you just need to add a parameter to the definition with type goapi.API_PARAM_TYPE_FILE and a
corresponding parameter to the handler of type goapi.File.
If the file is not required, make sure to call goapi.File.Valid() before using the file.
The handler owns the file once passed and is responsible for calling Close() on it.

#### Implicit Parameters

In addition to explicitly defined parameters, your handler can have up to five implicit parameters.
These parameters may appear in ANY order in your handler function and you may pick and choose the ones you want (or none).
 - userId goapi.UserId - The id of the user making the request (requires authentication).
 - userName goapi.UserName - The name of the user making the request (requires authentication).
 - token goapi.Token - The token of the user making the request (requires authentication).
 - request *http.Request - The http request.
 - response http.ResponseWriter - The http response (you should only use this in extreme cases).
Remember that in Go, we cannot get parameter names.
So you may call these parameters whatever you want, they are made unique by their types.
request and response are obvious, but userId, userName, and token are a little more unusual.
These are typed to be an int, string, and string respectively, and are only typed special to uniquely identify them.

### Handler Return Values

The return value for the handler is very flexible.
Up to three values can be returned:
 - interface{} - The value to be serialized and put in the http response.
                 This will usaully be turned to JSON.
                 Feel free to pass something like "" if you are also passing an error.
 - int - An http response code (eg http.StatusOK or http.StatusBadRequest).
         If 0, then the code will be inferred from the context.
 - error - Any error that occurred.
           In the case of an error, the response (interface{}) will be ignored and a failure response will be issued.
           The http status will still be honored.
Once again, you can specify anywhere between zero and all three return values.
The return values must be typed exactly.
However, they may be returned in any order.

## Security

Requests are authorized via tokens.
Tokens can be any string (such as any hex key or a JWT).
Tokens must be passed through "Authorization" HTTP header and be prefixed with "Bearer ".
Ex: "Bearer SOMETOKEN".

After picked up from the HTTP headers, tokens will be passed to the token validation function assigned to the ApiMethodFactory.
The id and name of the requesting user should be returned when a token is validated.
The id and name will be blindly passed to the handler (if requested by the handler), so feel free to default one if they are not used.

The signature for the validation function looks like:
```go
type ValidateToken func(token string, log Logger) (userId int, userName string, err error)
```

Authentication is controled on a per-ApiMethod basis using the auth parameter to ApiMethodFactory.NewApiMethod().
If turned off, there will be no attempt to fetch a token or validate tokens.

## Constructing ApiMethods

Use an ApiMethodFactory to construct ApiMethods (using ApiMethodFactory.NewApiMethod()).
An ApiMethodFactory has several configurable components:

### Content Type

The value for the "Content-Type" HTTP header.
Defaults to "application/json; charset=UTF-8".
Set using ApiMethodFactory.SetContentType().

### Logger

The logging mechanism to use while handling the API.
Defaults to a goapi.ConsoleLogger.
Set using ApiMethodFactory.SetLogger().

### Serializer

The serialization mechanism to use to convert handler responses into a string.
Defaults to a goapi.JSONSerializer.
Set using ApiMethodFactory.SetSerializer().

### Error Responder

A function to choose the proper response for errors while handling a request.
The function should return an object that can be handled by the factory's serializer.
Defaults to a goapi.GeneralErrorResponder.
Set using ApiMethodFactory.SetErrorResponder().

### Token Validation

The token validation mechanism to use while handling the API.
See the "Security" section for more information on the validation function.
If any ApiMethod uses authentication, then a validation method must be provided.
If a validation method is not provided and authentication is required, then ApiMethod validation will panic.
Set using ApiMethodFactory.SetTokenValidator().

## Validation

During construction, each ApiMethod will be validated.
Validation will check both the method definition, the parameter semantics, and return value semantics.
Validation will panic (actually call Panic() on the logger) on any error.
The logger should honor panic semantics and immediatly halt the runtime.

# Using an ApiMethod

After you get an ApiMethod, you can call ApiMethod.Middleware() to get a function that can be used by http.HandleFunc().

The returned function will have the signature:
```go
func(response http.ResponseWriter, request *http.Request)
```

Typical usage may look something like:
```go
import "net/http"
import "github.com/eriq-augustine/goapi"

func SetupAPI() {
   var factory ApiMethodFactory;
   factory.SetTokenValidator(myJWTTokenValidator);

   methods := []ApiMethod{
      factory.NewApiMethod(
         "do/something", // Path
         somethingHandler, // Handler
         false, // Do not authenticate
         []ApiMethodParam{
            {"someRequiredIntParam", goapi.API_PARAM_TYPE_STRING, true},
            {"someOptionalStringParam", goapi.API_PARAM_TYPE_STRING, false},
         },
      ),
      factory.NewApiMethod(
         "do/somethingElse",
         somethingElseHandler,
         true, // Authenticate
         []ApiMethodParam{}, // No params
      ),
   };

   // Register all the api methods.
   for _, method := range(methods) {
      http.HandleFunc(buildApiUrl(method.Path), method.Middleware());
   }
}
```
