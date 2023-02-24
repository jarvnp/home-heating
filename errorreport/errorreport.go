package errorreport

import(
  //"home-heating/email"
  "net/http"
  "time"
  "fmt"
)

const HEADER = "Lämmityssysteemin ERROR!"

var RECIPIENTS  = []string{"RECIPIENT_EMAIL"}

func Report(description string, errorText string){
  client := http.Client{
    Timeout: 60 * time.Second,
  }

  //while testing:
  fmt.Println("error: ", description + "\n\n" + errorText,client)

  /*err:= email.SendEmail(RECIPIENTS,HEADER,description + "\n\n" + errorText,client)
  if(err != nil){
    panic("Emailin lähetys ei onnistu:\n\n"+err.Error())
  }*/
}
