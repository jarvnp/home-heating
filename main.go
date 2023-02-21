package main

import(
  "fmt"
  "net/http"
  "time"
  "home-heating/electricityprice"
  "home-heating/shelly"
  "home-heating/temperature"
  //"home-heating/email"
  "home-heating/errorreport"
)

func main() {
  client := http.Client{
    Timeout: 60 * time.Second,
  }

  resp1, err1:= temperature.GetTemperature(client);
  if(err1 == nil){
    fmt.Println(resp1);
  }else{
    errorreport.Report("Error lämpötilan haussa", err1.Error())
  }

  err2:=shelly.SetSwitch(0,false, client);
  if(err2 != nil){
    errorreport.Report("Error shellyn kytkinten kanssa", err2.Error())
  }

  resp3,err3:=electricityprice.GetPrices("202302162300", "202302172300",client);
  if(err3 == nil){
    fmt.Println(resp3);
    fmt.Println(len(resp3))
  }else{

    errorreport.Report("Error pörssisähkön hintojen haussa", err3.Error())
  }
/*
  text := "test\\ntest"
  err4:=email.SendEmail([]string{"RECIPIENT_EMAIL", "RECIPIENT_EMAIL"},"TEST",text,client)
  fmt.Println(err4)*/
}
