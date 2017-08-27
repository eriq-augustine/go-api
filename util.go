package goapi;

import (
   "encoding/json"
   "net/http"
   "strings"
);

func getToken(request *http.Request, allowTokenParam bool) (string, bool) {
   var tokenText string = "";

   // First check the header and then the query params if allowed.
   authHeader, ok := request.Header["Authorization"];
   if (ok) {
      tokenText = authHeader[0];
   } else if (!ok && allowTokenParam) {
      request.ParseMultipartForm(MULTIPART_PARSE_SIZE);

      tokenText = request.FormValue(PARAM_TOKEN);
   }

   token := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(tokenText), "Bearer"));
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
