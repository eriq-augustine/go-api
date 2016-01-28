package goapi;

import (
   "fmt"
   "net/http"
   "testing"
)

type TestInfo struct {
   title string
   path string
   handler interface{}
   auth bool
   params []ApiMethodParam
   valid bool
}

func ExampleApiMethod_String_simple() {
   factory := ApiMethodFactory{};
   factory.SetTokenValidator(fakeValidateToken);
   method := factory.NewApiMethod("/good/path", handler_empty, true, []ApiMethodParam{});
   fmt.Println(method);

   // Output:
   // /good/path
   //    Authentication Required: true
   //    Params: None
   //    Handler: github.com/eriq-augustine/goapi.handler_empty
}

func ExampleApiMethod_String_params() {
   factory := ApiMethodFactory{};
   factory.SetTokenValidator(fakeValidateToken);
   method := factory.NewApiMethod(
      "/other/path",
      handler_intString,
      false,
      []ApiMethodParam{
         ApiMethodParam{"someInt", API_PARAM_TYPE_INT, true},
         ApiMethodParam{"someString", API_PARAM_TYPE_STRING, false},
      },
   );
   fmt.Println(method);

   // Output:
   // /other/path
   //    Authentication Required: false
   //    Params:
   //       someInt int (required)
   //       someString string
   //    Handler: github.com/eriq-augustine/goapi.handler_intString
}

func TestValidation(t *testing.T) {
   tests := []TestInfo{
      {
         title: "Valid - No Params",
         path: "/good/path",
         handler: handler_empty,
         auth: false,
         params: []ApiMethodParam{},
         valid: true,
      },
      {
         title: "Valid - Request",
         path: "/good/path",
         handler: handler_request,
         auth: false,
         params: []ApiMethodParam{},
         valid: true,
      },
      {
         title: "Valid - Response",
         path: "/good/path",
         handler: handler_response,
         auth: false,
         params: []ApiMethodParam{},
         valid: true,
      },
      {
         title: "Valid - Auth Empty",
         path: "/good/path",
         handler: handler_empty,
         auth: true,
         params: []ApiMethodParam{},
         valid: true,
      },
      {
         title: "Valid - Auth Special Types",
         path: "/good/path",
         handler: handler_implicits,
         auth: true,
         params: []ApiMethodParam{},
         valid: true,
      },
      {
         title: "Valid - Params String",
         path: "/good/path",
         handler: handler_string,
         auth: false,
         params: []ApiMethodParam{
            ApiMethodParam{"someString", API_PARAM_TYPE_STRING, true},
         },
         valid: true,
      },
      {
         title: "Valid - Params Int",
         path: "/good/path",
         handler: handler_int,
         auth: false,
         params: []ApiMethodParam{
            ApiMethodParam{"someInt", API_PARAM_TYPE_INT, true},
         },
         valid: true,
      },
      {
         title: "Valid - Params Multiple int/string",
         path: "/good/path",
         handler: handler_multipleIntString,
         auth: false,
         params: []ApiMethodParam{
            ApiMethodParam{"someString1", API_PARAM_TYPE_STRING, true},
            ApiMethodParam{"someInt1", API_PARAM_TYPE_INT, true},
            ApiMethodParam{"someString2", API_PARAM_TYPE_STRING, true},
            ApiMethodParam{"someInt2", API_PARAM_TYPE_INT, true},
         },
         valid: true,
      },
      {
         title: "Valid - Params All",
         path: "/good/path",
         handler: handler_all,
         auth: true,
         params: []ApiMethodParam{
            ApiMethodParam{"someString1", API_PARAM_TYPE_STRING, true},
            ApiMethodParam{"someInt1", API_PARAM_TYPE_INT, true},
            ApiMethodParam{"someString2", API_PARAM_TYPE_STRING, true},
            ApiMethodParam{"someInt2", API_PARAM_TYPE_INT, true},
         },
         valid: true,
      },
      {
         title: "Valid - Return 1",
         path: "/good/path",
         handler: handler_return1,
         auth: false,
         params: []ApiMethodParam{},
         valid: true,
      },
      {
         title: "Valid - Return 2",
         path: "/good/path",
         handler: handler_return2,
         auth: false,
         params: []ApiMethodParam{},
         valid: true,
      },
      {
         title: "Valid - Return 3",
         path: "/good/path",
         handler: handler_return3,
         auth: false,
         params: []ApiMethodParam{},
         valid: true,
      },
      // Invalid methods
      {
         title: "Invalid - Empty Path",
         path: "",
         handler: handler_empty,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
      {
         title: "Inalid - Empty Handler",
         path: "/good/path",
         handler: nil,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
      {
         title: "Inalid - Undeclared Param",
         path: "/good/path",
         handler: handler_int,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
      {
         title: "Inalid - Missing Param",
         path: "/good/path",
         handler: handler_empty,
         auth: false,
         params: []ApiMethodParam{
            ApiMethodParam{"someString", API_PARAM_TYPE_STRING, true},
         },
         valid: false,
      },
      {
         title: "Inalid - No Auth UserId",
         path: "/good/path",
         handler: handler_userId,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
      {
         title: "Inalid - No Auth UserName",
         path: "/good/path",
         handler: handler_userName,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
      {
         title: "Inalid - No Auth Token",
         path: "/good/path",
         handler: handler_token,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
      {
         title: "Inalid - Params Empty Name",
         path: "/good/path",
         handler: handler_empty,
         auth: false,
         params: []ApiMethodParam{
            ApiMethodParam{"", API_PARAM_TYPE_STRING, true},
         },
         valid: false,
      },
      {
         title: "Inalid - Params Bad Type",
         path: "/good/path",
         handler: handler_int,
         auth: false,
         params: []ApiMethodParam{
            ApiMethodParam{"someInt", 99, true},
         },
         valid: false,
      },
      {
         title: "Inalid - Dup Name",
         path: "/good/path",
         handler: handler_int,
         auth: false,
         params: []ApiMethodParam{
            ApiMethodParam{"someInt", API_PARAM_TYPE_INT, true},
            ApiMethodParam{"someInt", API_PARAM_TYPE_INT, true},
         },
         valid: false,
      },
      {
         title: "Inalid - Return 4",
         path: "/good/path",
         handler: handler_return4,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
      {
         title: "Inalid - Return Bad Types",
         path: "/good/path",
         handler: handler_returnBadType,
         auth: false,
         params: []ApiMethodParam{},
         valid: false,
      },
   };

   factory := ApiMethodFactory{};
   factory.SetTokenValidator(fakeValidateToken);
   for _, test := range(tests) {
      validationTest(t, factory, test);
   }
}

func validationTest(t *testing.T, factory ApiMethodFactory, info TestInfo) {
   defer func() {
      // Check panic status.
      r := recover();

      if (r != nil && info.valid) {
         // Too panicy.
         t.Errorf("%s: Invalid Panic", info.title)
      } else if (r == nil && !info.valid) {
         // Not panicy enough.
         t.Errorf("%s: Failed to Panic", info.title)
      }
   }();

   factory.NewApiMethod(info.path, info.handler, info.auth, info.params);
}

func fakeValidateToken(token string, log Logger) (userId int, userName string, err error) {
   return 0, "", nil;
}

// Test handlers.

func handler_empty() {
}

func handler_request(request *http.Request) {
}

func handler_response(response http.ResponseWriter) {
}

func handler_implicits(id UserId, name UserName, token Token) {
}

func handler_userId(id UserId) {
}

func handler_userName(name UserName) {
}

func handler_token(token Token) {
}

func handler_string(someString string) {
}

func handler_int(someInt int) {
}

func handler_intString(someInt int, someString string) {
}

func handler_multipleIntString(someInt1 int, someString1 string, someInt2 int, someString2 string) {
}

func handler_all(token Token, someInt1 int, someString1 string, id UserId, someInt2 int, someString2 string, name UserName) {
}

func handler_return1() (int) {
   return 0;
}

func handler_return2() (int, error) {
   return 0, nil;
}

func handler_return3() (error, int, interface{}) {
   return nil, 0, nil;
}

func handler_return4() (error, int, interface{}, int) {
   return nil, 0, nil, 0;
}

func handler_returnBadType() (bool) {
   return false;
}
