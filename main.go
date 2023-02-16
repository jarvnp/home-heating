package main

import(
  "fmt"
  "net/http"
  "time"
  "home-heating/shelly"
  "home-heating/temperature"
)

func main() {
  client := http.Client{
    Timeout: 10 * time.Second,
  }

  resp, err:= temperature.GetTemperature(client);
  if(err == nil){
    fmt.Println(resp);
  }

  err=shelly.SetSwitch(1,false, client);
  if(err != nil){
    fmt.Println(err);
  }
}
