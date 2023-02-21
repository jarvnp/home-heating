package main

import(
  "fmt"
  "net/http"
  "time"
  //"home-heating/electricityprice"
  //"home-heating/shelly"
  //"home-heating/temperature"
  "home-heating/email"
)

func main() {
  client := http.Client{
    Timeout: 10 * time.Second,
  }
/*
  resp1, err1:= temperature.GetTemperature(client);
  if(err1 == nil){
    fmt.Println(resp1);
  }else{
    fmt.Println("error: ",err1);
  }

  err2:=shelly.SetSwitch(0,false, client);
  if(err2 != nil){
    fmt.Println(err2);
  }

  resp3,err3:=electricityprice.GetPrices("202302162300", "202302172300",client);
  if(err3 == nil){
    fmt.Println(resp3);
    fmt.Println(len(resp3))
  }else{

    fmt.Println("error: ",err3);
  }*/

  text := "test\\ntest"
  err4:=email.SendEmail([]string{"RECIPIENT_EMAIL", "RECIPIENT_EMAIL"},"TEST",text,client)
  fmt.Println(err4)
}
