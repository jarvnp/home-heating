package main

import(
  "fmt"
  "home-heating/errorreport"
  "home-heating/limitplan"
  "home-heating/shelly"
  "home-heating/config"
  "home-heating/jsonrw"
  "os"
)




func main() {

  var plan limitplan.PlanData;

  err := jsonrw.ReadFromJsonFile("plan.json", &plan)
  if(err != nil){
    errorreport.Report("Error plan tiedoston avaamisessa", err.Error(), config.ERROR_CODE_JSON)
    os.Exit(0)
  }


  err = limitplan.UpdatePlan("plan.json",&plan)
  if(err != nil){
    errorreport.Report("Error tietojen haussa", err.Error(), config.ERROR_CODE_DATA_FETCH)
    os.Exit(0)
  }



  limit,err := limitplan.FindLimitForNow(&plan)
  if(err != nil){
    errorreport.Report("Ongelmia rajoituksen haussa", err.Error(), config.ERROR_CODE_LIMIT_CALC)
    os.Exit(0)
  }
  fmt.Println(limit)

  err = shelly.SetLimit(limit)
  if(err != nil){
    errorreport.Report("Ongelmia shellyn kanssa", err.Error(), config.ERROR_CODE_SHELLY)
    os.Exit(0)
  }
}
