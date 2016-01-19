package goapi;

import (
   "encoding/json"
   "net/http"
   "strings"
);

func getToken(request *http.Request) (string, bool) {
   authHeader, ok := request.Header["Authorization"];

   if (!ok) {
      return "", false;
   }

   if (len(authHeader) == 0) {
      return "", false;
   }

   token := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(authHeader[0]), "Bearer"));

   if (token == "") {
      return "", false;
   }

   return token, true;
}

func toJSON(data interface{}) (string, error) {
   bytes, err := json.Marshal(data);
   if (err != nil) {
      return "", err;
   }

   return string(bytes), nil;
}
