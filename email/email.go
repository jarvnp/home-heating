package email

import(
  "net/http"
  "strings"
  "errors"
  "io/ioutil"
  "encoding/json"
  "fmt"
)

const URL = "https://send.api.mailtrap.io/api/send"


const AUTH = "EMAIL_TOKEN"
const FROM_EMAIL = "FROM_EMAIL"
const FROM_NAME = "Lämmityssysteemi"


type Response struct{
  Success bool `json:"success"`
  Errors []string `json:errors`
}


func getMessageBody(recipients []string, subject string, message string)string{
  body :=`{
    "from":{"email":"%s","name":"%s"},
    "to":[%s],
    "subject":"%s",
    "text":"%s"
  }`

  var recipientsStr string;
  for i:= range recipients{
    if(i != 0){
      recipientsStr += ","
    }
    recipientsStr += fmt.Sprintf("{\"email\":\"%s\"}", recipients[i])
  }

  body = fmt.Sprintf(body, FROM_EMAIL,FROM_NAME,recipientsStr,subject,message)
  fmt.Println(body)
  return body
}


func SendEmail(recipients []string, subject string, message string, client http.Client)error{
  req, err := http.NewRequest("POST", URL, strings.NewReader(getMessageBody(recipients,subject,message)));
  if(err != nil){
    return err;
  }
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Api-Token", AUTH)

  resp, err := client.Do(req)
  if(err != nil){
    return err;
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err;
  }
  var dat Response;
  if err := json.Unmarshal(body, &dat); err != nil {
      return err
  }
  if(!dat.Success){
    var errorText string;
    for i:=range dat.Errors{
      if(i != 0){
        errorText += ",\n"
      }
      errorText += dat.Errors[i];
    }
    return errors.New(errorText);
  }
  if(resp.StatusCode != 200){
    return errors.New(resp.Status);
  }
  defer resp.Body.Close()
  return nil;
}
