package main

import(
  "fmt"
  "home-heating/errorreport"
  "home-heating/limitplan"
  "os"
)




func main() {

  var plan limitplan.PlanData;

  err := limitplan.GetPlansFromJson("plan.json", &plan)
  if(err != nil){
    errorreport.Report("Error plan tiedoston avaamisessa", err.Error())
    os.Exit(0)
  }


  err = limitplan.UpdatePlan("plan.json",&plan)
  if(err != nil){
    errorreport.Report("Error tietojen haussa", err.Error())
    os.Exit(0)
  }



  limit,err := limitplan.FindLimitForNow(&plan)
  if(err != nil){
    errorreport.Report("Ongelmia rajoituksen haussa", err.Error())
    os.Exit(0)
  }
  fmt.Println(limit)
/*
  var temp float64
  var prices []float64

  if(!getDataAndReportError(&temp, &prices)){
    os.Exit(0)
  }
  fmt.Println(temp)
  fmt.Println(prices)

  newPlan, _ := dayplan.Plan(10.0,prices);
  fmt.Println(newPlan)*/
}
