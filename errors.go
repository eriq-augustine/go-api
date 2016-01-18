package goapi;

const (
   TOKEN_VALIDATION_NO_TOKEN = iota
   TOKEN_VALIDATION_EXPIRED
   TOKEN_VALIDATION_REVOKED
   TOKEN_VALIDATION_BAD_SIGNATURE
   TOKEN_AUTH_BAD_CREDENTIALS
);

type TokenValidationError struct {
   Reason int
};

func (err TokenValidationError) Description() string {
   switch err.Reason {
   case TOKEN_VALIDATION_NO_TOKEN:
      return "Token does not exist";
   case TOKEN_VALIDATION_EXPIRED:
      return "Token is expired";
   case TOKEN_VALIDATION_REVOKED:
      return "Token has been revoked";
   case TOKEN_VALIDATION_BAD_SIGNATURE:
      return "Token failed integrity check";
   case TOKEN_AUTH_BAD_CREDENTIALS:
      return "Bad credentials";
   default:
      return "Unknown token error";
   }
}

func (err TokenValidationError) Error() string {
   return err.Description();
}
