package email

import (
	"encoding/json"
	"errors"
	"home-heating/secret"
	"io/ioutil"
	"net/http"
	"strings"
)

const URL = "https://send.api.mailtrap.io/api/send"






type Response struct{
  Success bool `json:"success"`
  Errors []string `json:"errors"`
}

type Recipient struct{
  Email string  `json:"email"`
}

type Request struct{
  From struct{
    Email string  `json:"email"`
    Name string `json:"name"`
  }`json:"from"`

  To []Recipient `json:"to"`

  Subject string  `json:"subject"`
  Text string `json:"text"`
}

func getMessageBody(recipients []string, subject string, message string)string{
  var body Request;
  body.Subject = subject
  body.Text = message
  body.From.Email = secret.FROM_EMAIL
  body.From.Name = secret.FROM_NAME

  for i:= range recipients{
    var recipient Recipient;
    recipient.Email = recipients[i]
    body.To = append(body.To, recipient)
  }


  bodyStr,_ := json.Marshal(&body)
  return string(bodyStr)
}


func SendEmail(recipients []string, subject string, message string, client http.Client)error{
  req, err := http.NewRequest("POST", URL, strings.NewReader(getMessageBody(recipients,subject,message)));
  if(err != nil){
    return err;
  }
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Api-Token", secret.EMAIL_TOKEN)

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
