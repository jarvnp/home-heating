package main

import(
  "fmt"
  "time"
  "home-heating/errorreport"
  "home-heating/limitplan"
  "home-heating/shelly"
  "home-heating/config"
  "home-heating/jsonrw"
  "os"
)




func main() {

  var plan []limitplan.PlanData;

  err := jsonrw.ReadFromJsonFile("plan.json", &plan)
  if(err != nil){
    shelly.SetLimit(config.NO_LIMIT)
    errorreport.Report("Error plan tiedoston avaamisessa", err.Error(), config.ERROR_CODE_JSON)
    os.Exit(0)
  }


  err = limitplan.UpdatePlan("plan.json",&plan)
  if(err != nil){
    shelly.SetLimit(config.NO_LIMIT)
    errorreport.Report("Error tietojen haussa", err.Error(), config.ERROR_CODE_DATA_FETCH)
    os.Exit(0)
  }



  limit,err := limitplan.FindLimitForNow(&plan)
  if(err != nil){
    shelly.SetLimit(config.NO_LIMIT)
    errorreport.Report("Ongelmia rajoituksen haussa", err.Error(), config.ERROR_CODE_LIMIT_CALC)
    os.Exit(0)
  }
  fmt.Println(limit)


  // Try to reach shelly multiple times
  // This will hopefully reduce the amount of error messages sent
  // So far connecting to Shelly causes most of the error messages
  for i:=0;i<5;i++{
    err = shelly.SetLimit(limit)
    if(err == nil){
      break
    }
    fmt.Println("Trying to connect to shelly, i = ",i);
    fmt.Println("Error = ", err.Error());
    time.Sleep(30 * time.Second)
  }

  if(err != nil){
    errorreport.Report("Ongelmia shellyn kanssa", err.Error(), config.ERROR_CODE_SHELLY)
    os.Exit(0)
  }

  //no errors have occured if we've made it this far
  //if there has recently been an error, and now we are back at normal operation, the program reports that
  if(errorreport.IsRecovery()){
    errorreport.ReportRecovery()
    errorreport.ClearErrorHistory()
  }

}
