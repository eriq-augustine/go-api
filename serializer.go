package goapi;

type Serializer func(responseObj interface{}) (string, error)

func JSONSerializer(responseObj interface{}) (string, error) {
   return ToJSON(responseObj);
}
