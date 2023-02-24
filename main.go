package main

import(
  "fmt"
  "net/http"
  "time"
  "home-heating/electricityprice"
  //"home-heating/shelly"
  "home-heating/temperature"
  "home-heating/errorreport"
  "home-heating/dayplan"
  "os"
)

func main() {


  var temp float64
  var prices []float64

  if(!getDataAndReportError(&temp, &prices)){
    os.Exit(0)
  }
  fmt.Println(temp)
  fmt.Println(prices)

  plan, _ := dayplan.Plan(10.0,prices);
  fmt.Println(plan)
}


func getDataAndReportError(temp *float64, prices *[]float64)bool{
  client := http.Client{
    Timeout: 60 * time.Second,
  }
  var err error
  *temp, err = temperature.GetTemperature(client);
  if err != nil{
    errorreport.Report("Error lämpötilan haussa", err.Error())
    return false
  }
/*
  err=shelly.SetSwitch(0,false, client);
  if(err != nil){
    errorreport.Report("Error shellyn kytkinten kanssa", err.Error())
    return false
  }*/

  *prices,err = electricityprice.GetPrices("202206262300", "202206272300",client);
  if(err != nil){
    errorreport.Report("Error pörssisähkön hintojen haussa", err.Error())
    return false
  }

  return true
}
