package goapi;

import (
   "net/http"
   "testing"
);

func failTest(t *testing.T, title string, expected interface{}, actual interface{}) {
   t.Errorf("%s: Expected: %v, Got: %v", title, expected, actual);
}

func TestGetToken_Basic(t *testing.T) {
   tests := []struct{
      title string
      authHeaderKey string
      authHeaderValue string
      token string
      ok bool
   } {
      {
         title: "Basic",
         authHeaderKey: "Authorization",
         authHeaderValue: "Bearer TOKEN",
         token: "TOKEN",
         ok: true,
      },
      {
         title: "Bad Header",
         authHeaderKey: "BadHeader",
         authHeaderValue: "Bearer TOKEN",
         token: "TOKEN",
         ok: false,
      },
      {
         title: "Empty Value",
         authHeaderKey: "Authorization",
         authHeaderValue: "",
         token: "",
         ok: false,
      },
      {
         title: "Empty Token",
         authHeaderKey: "Authorization",
         authHeaderValue: "Bearer ",
         token: "",
         ok: false,
      },
      {
         title: "Whitspace Token",
         authHeaderKey: "Authorization",
         authHeaderValue: "Bearer            ",
         token: "",
         ok: false,
      },
      {
         title: "Trim Token",
         authHeaderKey: "Authorization",
         authHeaderValue: "     Bearer TOKEN    ",
         token: "TOKEN",
         ok: true,
      },
      {
         title: "Spaced Token",
         authHeaderKey: "Authorization",
         authHeaderValue: "Bearer T O K E N",
         token: "T O K E N",
         ok: true,
      },
      {
         title: "Base64 Token",
         authHeaderKey: "Authorization",
         authHeaderValue: "Bearer ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=",
         token: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=",
         ok: true,
      },
   };

   for _, test := range(tests) {
      request, err := http.NewRequest("GET", "http://example.com", nil);
      if (err != nil) {
         t.Error("Failed to create a request: ", err);
      }

      request.Header.Set(test.authHeaderKey, test.authHeaderValue);
      token, ok := getToken(request);

      if (ok != test.ok) {
         failTest(t, test.title, test.ok, ok);
         continue;
      }

      if (!ok) {
         continue;
      }

      if (token != test.token) {
         failTest(t, test.title, test.token, token);
         continue;
      }
   }
}
