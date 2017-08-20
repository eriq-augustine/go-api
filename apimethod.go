package goapi;

import (
   "fmt"
   "io"
   "io/ioutil"
   "mime/multipart"
   "net/http"
   "reflect"
   "runtime"
   "strconv"
   "strings"
)

const (
   MULTIPART_PARSE_SIZE = 4 * 1024 * 1024 // Store the first 4M in memory.
)

const (
   API_PARAM_TYPE_INT = iota
   API_PARAM_TYPE_STRING
   API_PARAM_TYPE_FILE
)

// We need to define these as types, so we can figure it out when we want to pass params.
type Token string;
type UserId int;
type UserName string;

type File struct {
   Io *multipart.File
}

func (file File) Valid() bool {
   return file.Io != nil;
}

func (file File) Data() ([]byte, error) {
   if (!file.Valid()) {
      return nil, fmt.Errorf("Cannot get data of invalid file");
   }

   return ioutil.ReadAll(*file.Io);
}

type ApiMethod struct {
   path string
   handler interface{}
   auth bool
   params []ApiMethodParam
   log Logger
   serializer Serializer
   contentType string
   errorResponder ErrorResponder
   tokenValidator ValidateToken
}

type ApiMethodParam struct {
   Name string
   ParamType int
   Required bool
}

func (method ApiMethod) Path() string {
   return method.path;
}

// Will just panic on error.
func (method ApiMethod) validate() {
   // Check the definitions.
   if (method.path == "") {
      method.log.Panic("Bad path for API handler");
   }

   if (method.handler == nil) {
      method.log.Panic(fmt.Sprintf("Nil handler for API handler for path: %s", method.path));
   }

   for _, param := range(method.params) {
      if (param.Name == "") {
         method.log.Panic(fmt.Sprintf("Empty name for param for API handler for path: %s", method.path));
      }

      if (!(param.ParamType == API_PARAM_TYPE_INT || param.ParamType == API_PARAM_TYPE_STRING || param.ParamType == API_PARAM_TYPE_FILE)) {
         method.log.Panic(fmt.Sprintf("Param (%s) for API handler (%s) has bad type (%d)", param.Name, method.path, param.ParamType));
      }
   }

   // Check parameter semantics.

   var handlerType reflect.Type = reflect.TypeOf(method.handler);

   var numParams int = handlerType.NumIn();
   var additionalParams = 0;

   for i := 0; i < numParams; i++ {
      var ParamType reflect.Type = handlerType.In(i);

      if (ParamType.String() == "goapi.Token") {
         additionalParams++;

         if (!method.auth) {
            method.log.Panic(fmt.Sprintf("API handler (%s) requested a token without authentication", method.path));
         }
      } else if (ParamType.String() == "goapi.UserId") {
         additionalParams++;

         if (!method.auth) {
            method.log.Panic(fmt.Sprintf("API handler (%s) requested a user id without authentication", method.path));
         }
      } else if (ParamType.String() == "goapi.UserName") {
         additionalParams++;

         if (!method.auth) {
            method.log.Panic(fmt.Sprintf("API handler (%s) requested a user name without authentication", method.path));
         }
      } else if (ParamType.String() == "*http.Request") {
         additionalParams++;
      } else if (ParamType.String() == "http.ResponseWriter") {
         additionalParams++;
      } else {
         if (!(ParamType.String() == "int" || ParamType.String() == "string" || ParamType.String() == "goapi.File")) {
            method.log.Panic(fmt.Sprintf("API handler (%s) has an actual parameter with incorrect type (%s) must be string or int", method.path, ParamType.String()));
         }
      }
   }

   if (numParams != len(method.params) + additionalParams) {
      method.log.Panic(fmt.Sprintf("API handler (%s) actually expects %d parameters, but is defined to expect %d (%d defined, %d implicit)", method.path, numParams, len(method.params) + additionalParams, len(method.params), additionalParams));
   }

   // Check the return semantics.
   var numReturns int = handlerType.NumOut();

   if (numReturns > 4) {
      method.log.Panic(fmt.Sprintf("API handler (%s) has too many return values. Got %d. Maximum is 4.", method.path, numReturns));
   }

   // All of the return types are unique.
   var seenReturnTypes map[string]bool = make(map[string]bool);
   for i := 0; i < numReturns; i++ {
      var returnType string = handlerType.Out(i).String();

      if (!(returnType == "interface {}" || returnType == "int" || returnType == "string" || returnType == "error")) {
         method.log.Panic(fmt.Sprintf("API handler (%s) has an bad return type (%s) must be interface{}, int, string, or error", method.path, returnType));
      }

      if (seenReturnTypes[returnType]) {
         method.log.Panic(fmt.Sprintf("API handler (%s) has duplicate return types (%s). Can only have up to four return types and each must be a unique type (interface{}, int, string, or error).", method.path, returnType));
      }
      seenReturnTypes[returnType] = true;
   }
}

func (method ApiMethod) Middleware() func(response http.ResponseWriter, request *http.Request) {
   return func(response http.ResponseWriter, request *http.Request) {
      response.Header().Set("Access-Control-Allow-Origin", "*");
      response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS");
      response.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization");

      // Skip preflight checks.
      if (request.Method == "OPTIONS") {
         return;
      }

      if (request.URL != nil) {
         method.log.Debug(request.URL.String());
      }

      responseObj, httpStatus, contentType, err := method.handleAPIRequest(response, request);

      // Wait until the handler responds to call set the content type.
      response.Header().Set("Content-Type", contentType);

      // TODO(eriq): Cleanup use of sendresponse.
      //  Directly call for an error, maybe.
      if (err != nil) {
         method.sendResponse("", err, httpStatus, response);
         return;
      }

      // If the response object is a reader, then stream it into the response writer.
      reader, ok := responseObj.(io.Reader);
      if (ok) {
         if (httpStatus == 0) {
            httpStatus = http.StatusOK;
         }

         response.WriteHeader(httpStatus);
         _, err = io.Copy(response, reader);
         if (err != nil) {
            // The reeponse may have got partially written... so just abandon the request.
            method.log.ErrorE("Failed to stream the response", err);
            return;
         }

         // Check to see if we need to close the reader.
         closer, ok := reader.(io.Closer);
         if (ok) {
            err = closer.Close();
            if (err != nil) {
               method.log.WarnE("Error closing a response reader, but the response still went out fine.", err);
            }
         }
      } else {
         // Otherwise, just serialize the response and send it over.
         responseString, err := method.serializer(responseObj);
         method.sendResponse(responseString, err, httpStatus, response);
      }
   }
}

// This handles the API side of the request.
// None of the boilerplate.
func (method ApiMethod) handleAPIRequest(response http.ResponseWriter, request *http.Request) (interface{}, int, string, error) {
   var userId int = -1;
   var userName string = "";
   var ok bool;
   var token string = "";

   if (method.auth) {
      var responseObject interface{};
      ok, userId, userName, token, responseObject = method.authRequest(request);
      if (!ok) {
         return responseObject, http.StatusUnauthorized, method.contentType, nil;
      }
   }

   ok, args := method.createArguments(UserId(userId), UserName(userName), Token(token), response, request);
   if (!ok) {
      return method.errorResponder(nil, http.StatusBadRequest), http.StatusBadRequest, method.contentType, nil;
   }

   var handlerValue reflect.Value = reflect.ValueOf(method.handler);
   returns := handlerValue.Call(args);

   return method.createReturnValues(returns);
}

func (method ApiMethod) createReturnValues(returns []reflect.Value) (interface{}, int, string, error) {
   var responseObj interface{} = nil;
   var httpStatus int = 0;
   var err error = nil;
   var contentType string = method.contentType;

   // Returns are optional.
   for _, val := range(returns) {
      var returnType reflect.Type = val.Type();

      if (returnType.String() == "interface {}") {
         if (!val.IsNil()) {
            responseObj = val.Interface();
         }
      } else if (returnType.String() == "int") {
         httpStatus = int(val.Int());
      } else if (returnType.String() == "string") {
         if (val.String() != "") {
            contentType = val.String();
         }
      } else if (returnType.String() == "error") {
         if (!val.IsNil()) {
            err = val.Interface().(error);
         }
      } else {
         method.log.Fatal(fmt.Sprintf("Unkown return type (%s) for API handler for path: %s", returnType.String(), method.path));
      }
   }

   return responseObj, httpStatus, contentType, err;
}

// Get all the parameters setup for invocation.
func (method ApiMethod) createArguments(userId UserId, userName UserName, token Token, response http.ResponseWriter, request *http.Request) (bool, []reflect.Value) {
   var handlerType reflect.Type = reflect.TypeOf(method.handler);
   var numParams int = handlerType.NumIn();

   var apiParamIndex = 0;
   var paramValues []reflect.Value = make([]reflect.Value, numParams);

   for i := 0; i < numParams; i++ {
      var ParamType reflect.Type = handlerType.In(i);

      // The user id, token, request, and response get handled specially.
      if (method.auth && ParamType.String() == "goapi.Token") {
         paramValues[i] = reflect.ValueOf(token);
      } else if (method.auth && ParamType.String() == "goapi.UserId") {
         paramValues[i] = reflect.ValueOf(userId);
      } else if (method.auth && ParamType.String() == "goapi.UserName") {
         paramValues[i] = reflect.ValueOf(userName);
      } else if (ParamType.String() == "*http.Request") {
         paramValues[i] = reflect.ValueOf(request);
      } else if (ParamType.String() == "http.ResponseWriter") {
         paramValues[i] = reflect.ValueOf(response);
      } else {
         // Normal param, fetch the next api parameter and pass it along.
         ok, val := method.fetchParam(apiParamIndex, request);
         if (!ok) {
            return false, []reflect.Value{};
         }

         paramValues[i] = val;
         apiParamIndex++;
      }
   }

   return true, paramValues;
}

func (method ApiMethod) fetchParam(apiParamIndex int, request *http.Request) (bool, reflect.Value) {
   // Only the first call will do anything.
   request.ParseMultipartForm(MULTIPART_PARSE_SIZE);

   var param ApiMethodParam = method.params[apiParamIndex];

   if (param.ParamType == API_PARAM_TYPE_FILE) {
      file, _, err := request.FormFile(param.Name);
      if (err != nil) {
         if (param.Required) {
            method.log.Warn(fmt.Sprintf("Required file parameter not found: %s", param.Name));
            return false, reflect.Value{};
         } else {
            return true, reflect.ValueOf(File{nil});
         }
      }

      return true, reflect.ValueOf(File{&file});
   }

   var stringValue string = strings.TrimSpace(request.FormValue(param.Name));

   if (param.Required && stringValue == "") {
      method.log.Warn(fmt.Sprintf("Required parameter not found: %s", param.Name));
      return false, reflect.Value{};
   }

   // If we are looking for string, then we are done.
   if (param.ParamType == API_PARAM_TYPE_STRING) {
      return true, reflect.ValueOf(stringValue);
   }

   // We must be looking for an int (only ints, files, and strings are allowed).

   // First check for an empty non-required int.
   if (stringValue == "") {
      return true, reflect.ValueOf(0);
   }

   intValue, err := strconv.Atoi(stringValue);
   if (err != nil) {
      method.log.WarnE(fmt.Sprintf("Unable to convert int parameter (%s) from string: '%s'", param.Name, stringValue), err);
      return false, reflect.ValueOf(0);
   }

   return true, reflect.ValueOf(intValue);
}

// Send a response over |response|.
// On error, |responseString| will be ignored.
// In not supplied, the |httpStatus| will become http.StatusInternalServerError on error and
// http.StatusOK on success.
func (method ApiMethod) sendResponse(responseString string, err error, httpStatus int, response http.ResponseWriter) {
   if (err != nil) {
      method.log.ErrorE("API Error", err);

      if (httpStatus == 0) {
         httpStatus = http.StatusInternalServerError;
      }

      // Any serialization errors will be ignored at this point.
      responseString, _ = method.serializer(method.errorResponder(err, httpStatus));
      response.WriteHeader(httpStatus);
      fmt.Fprintln(response, responseString);
   } else {
      method.log.Debug("Successful Response:\n" + responseString);

      if (httpStatus == 0) {
         httpStatus = http.StatusOK;
      }

      response.WriteHeader(httpStatus)
      fmt.Fprintln(response, responseString);
   }
}

// Tries to authorize a request.
// Returns: success, user id, user name, request token, and response object.
// user id and token will only be populated on success.
// response object will only be populated on error.
func (method ApiMethod) authRequest(request *http.Request) (bool, int, string, string, interface{}) {
   token, ok := getToken(request);

   if (!ok) {
      return false, 0, "", "", method.errorResponder(TokenValidationError{TOKEN_VALIDATION_NO_TOKEN}, http.StatusUnauthorized);
   }

   // Check for empty tokens.
   if (strings.TrimSpace(token) == "") {
      return false, 0, "", "", method.errorResponder(TokenValidationError{TOKEN_VALIDATION_NO_TOKEN}, http.StatusUnauthorized);
   }

   userId, userName, err := method.tokenValidator(token, method.log);
   if (err != nil) {
      validationErr, ok := err.(TokenValidationError);
      if (!ok) {
         // Some other (non-validation) error.
         return false, 0, "", "", method.errorResponder(nil, http.StatusInternalServerError);
      }

      return false, 0, "", "", method.errorResponder(validationErr, http.StatusUnauthorized);
   }

   return true, userId, userName, token, nil;
}

func (method ApiMethod) String() string {
   var rtn string = "";

   rtn += fmt.Sprintf("%s\n", method.path);
   rtn += fmt.Sprintf("   Authentication Required: %v\n", method.auth);

   if (len(method.params) == 0) {
      rtn += "   Params: None\n"
   } else {
      rtn += "   Params:\n"
      for _, param := range(method.params) {
         rtn += fmt.Sprintf("      %v\n", param);
      }
   }

   rtn += fmt.Sprintf("   Handler: %s ", runtime.FuncForPC(reflect.ValueOf(method.handler).Pointer()).Name());
   return rtn;
}

func (param ApiMethodParam) String() string {
   var typeString string = "int";
   if (param.ParamType == API_PARAM_TYPE_STRING) {
      typeString = "string";
   } else if (param.ParamType == API_PARAM_TYPE_FILE) {
      typeString = "File";
   }

   var requiredString string = "";
   if (param.Required) {
      requiredString = " (required)";
   }

   return fmt.Sprintf("%s %s%s", param.Name, typeString, requiredString);
}
