package errorreport

import(
  //"home-heating/email"
  "net/http"
  "time"
  "fmt"
)

const HEADER = "Lämmityssysteemin ERROR!"



func Report(description string, errorText string){
  client := http.Client{
    Timeout: 60 * time.Second,
  }

  //while testing:
  fmt.Println("error: ", description + "\n\n" + errorText,client)

  /*err:= email.SendEmail(config.RECIPIENTS,HEADER,description + "\n\n" + errorText,client)
  if(err != nil){
    panic("Emailin lähetys ei onnistu:\n\n"+err.Error())
  }*/
}
