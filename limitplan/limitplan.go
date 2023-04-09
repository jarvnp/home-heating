package limitplan
import(
  "errors"
  "fmt"
  "sort"
  "time"
  "net/http"
  "home-heating/electricityprice"
  "home-heating/temperature"
  "home-heating/config"
  "home-heating/jsonrw"
)


type PlanData struct{
  Time string //format: ddmmyyyyhhmm
  Limit int
}



const MINIMI_TEMP = (24*config.MAKSIMI_TEHO-101.51)/(-4.59)  //Missä lämpötilassa lämmitys pitää olla koko ajan päällä. Perustuu energyUsage-function tietoihin.




//Palauttaa arvioidun sähkönkäytön vuorokaudessa lämpötilan mukaan. [t] = Celcius
//[E] = kWh
func energyUsage(temperature float64)float64{
  return -4.59*temperature+101.51
}

//Palauttaa sen ajan, jolloin lämmitystä rajoitetaan, lämpötilan funktiona.
//lämpötilan yksikkö Celcius, ajan yksikkö tunti
func passiveTime(temperature float64)int{
  //t = E/P
  return int(24.0-energyUsage(temperature)/plannedPowerWhenActive(temperature))
}

func plannedPowerWhenActive(temperature float64)float64{
  var TEHO_KUN_PAKKASTA_10 float64= energyUsage(-10)/(24-config.PASSIIVISET_TUNNIT_KUN_PAKKASTA_10)

  //lineaarinen sovitus tavoitellulle teholle siten käyttäen arvoja (MINIMI_TEMP,MAKSIMITEHO), (-10,TEHO_KUN_PAKKASTA_10)
  return ((config.MAKSIMI_TEHO-TEHO_KUN_PAKKASTA_10)/(MINIMI_TEMP+10))*(temperature-MINIMI_TEMP)+config.MAKSIMI_TEHO;
}


func getLimits(temperature float64, prices []float64, plan *[]PlanData, fetchPeriodStartDate time.Time)([]int,error){
  //the linear models don't apply to too high temperatures
  if(temperature > 15.0){
    temperature = 15.0
  }


  //this *should* work even with daylight savings time, when there sometimes are 23 or 25 hours in a day.
  //This is because the prices are fetched using UTC time, which is unaffected by daylight savings
  if(len(prices) != 24){
    return nil, errors.New("Error with price data amount: " + fmt.Sprint(prices))
  }


  type priceAndIndex struct{
      Price float64
      Index int
  }
  var pricesWithIndex []priceAndIndex;
  for i:= range prices{
    var a priceAndIndex;
    a.Price = prices[i]
    a.Index = i
    pricesWithIndex = append(pricesWithIndex,a)
  }

  sort.Slice(pricesWithIndex, func(i,j int)bool{
    return pricesWithIndex[i].Price > pricesWithIndex[j].Price
  })
  //fmt.Println(pricesWithIndex)
  var plannedLimits []int
  for i:=0;i<24;i++{
    plannedLimits = append(plannedLimits,config.NO_LIMIT)
  }

  var passiveTime = passiveTime(temperature)

  for i:=0; i<passiveTime; i++{
    plannedLimits[pricesWithIndex[i].Index] = config.TOTAL_LIMIT;
  }

  err:=  addBuffer(&plannedLimits, plan, fetchPeriodStartDate)

  return plannedLimits,err

}


//after total limit we will limit the following hours with small limit to prevent power surge after TOTAL_LIMIT
//Check also if yesterday ended with full limit (or unfinished small limit),
//in which case today must begin with small limit
func addBuffer(plannedLimits *[]int, plan *[]PlanData, fetchPeriodStartDate time.Time)error{

  var totalLimitHoursSequental int = 0

  var yesterdaysSmallLimitHours int = 0;

  var startIndex = len(*plan)-1

  //check if the most recent plan is from yesterday

  if(startIndex >= 0){
    lastPlanHourDate,err := timeStrToTime((*plan)[startIndex].Time)
    if(err != nil){
      return err;
    }

    if(fetchPeriodStartDate.Sub(lastPlanHourDate) <= time.Hour){


      //if the most recent plan is from yesterday, check if yesterday ended with full limit (or unfinished small limit),
      //in which case today must begin with small limit

      for i:=startIndex; i>=0; i--{
        if((*plan)[i].Limit == config.TOTAL_LIMIT){
          totalLimitHoursSequental++
        }else if( ((*plan)[i].Limit == config.SMALL_LIMIT) && (totalLimitHoursSequental == 0) ){
          yesterdaysSmallLimitHours++
        }else{
          break
        }
      }

      //after this totalLimitHoursSequental will correspond to the hours that need to be limited in the beginning of this new day
      if(totalLimitHoursSequental > config.MAX_LIMIT_HOURS_AFTER_TOTAL_LIMIT){
        totalLimitHoursSequental = config.MAX_LIMIT_HOURS_AFTER_TOTAL_LIMIT
      }
      totalLimitHoursSequental -= yesterdaysSmallLimitHours

      //shoudn't happen
      if(totalLimitHoursSequental < 0){
        fmt.Println("unexpected: totalLimitHoursSequental < 0")
        totalLimitHoursSequental = 0
      }
    }
  }



  for i:= range *plannedLimits{
    if((*plannedLimits)[i] == config.TOTAL_LIMIT){
      totalLimitHoursSequental++
    }else{
      if(totalLimitHoursSequental > 0){
        limitTime := 0

        //how many not limited hours there are after total limit
        //we will limit at max the same time that the total limit lasted
        for j:=0; j<totalLimitHoursSequental && j+i<len(*plannedLimits) && (*plannedLimits)[j+i] == config.NO_LIMIT; j++{
          limitTime++
        }

        if(limitTime > config.MAX_LIMIT_HOURS_AFTER_TOTAL_LIMIT){
          limitTime = config.MAX_LIMIT_HOURS_AFTER_TOTAL_LIMIT
        }

        for j:=0; j<limitTime; j++{
          (*plannedLimits)[j+i] = config.SMALL_LIMIT
        }

        i+=limitTime
      }

      totalLimitHoursSequental = 0
    }
  }
  return nil
}



func timeStrToTime(timeStr string)(time.Time,error){
  time,err := time.Parse("020120061504", timeStr)
  return time,err
}



func UpdatePlan(filename string, plan *[]PlanData)error{

  //remove old data from plan
  for len(*plan) > 0{
    planTime, err := timeStrToTime((*plan)[0].Time)
    if(err != nil){
      return err
    }
    if(time.Now().UTC().Sub(planTime) > time.Hour*24*config.PLAN_STORE_DURATION){
      *plan = (*plan)[1:]
    }else{
      break
    }
  }
  jsonrw.WriteToJsonFile(filename,plan)


  todayString := time.Now().UTC().Format("02012006")
  tomorrowString := time.Now().UTC().AddDate(0,0,1).Format("02012006")


  var lastPlanHourDate string = ""
  if(len(*plan) != 0){
    lastPlanHourDate = (*plan)[len(*plan)-1].Time

    //remove 4 last characters (clock information)
    lastPlanHourDate = lastPlanHourDate[0:len(lastPlanHourDate)-4]
  }

  var haveTomorrow bool = true
  var haveToday bool = true
  if(lastPlanHourDate != tomorrowString){
    haveTomorrow = false
    if(lastPlanHourDate != todayString){
      haveToday = false
    }
  }

  //if we already have today's data, we don't need to try to fetch tomorrow's data until afternoon
  //(the data isn't even published earlier)
  if(haveToday && time.Now().UTC().Hour() < 14){
    return nil
  }

  if(!haveToday || !haveTomorrow){



    //If we fect tomorror's data, the fetch period starts today. This is because the price API considers days to begin at 2300 (or 2200 during summertime)
    //we request tomorrows date beginning from today 2200 and ending tomorrow 2200
    //This is because our API considers days to start (in Finland) at 2300 or 2200 depending on weather we live in summer time or winter time
    //The last hours of a day are published with the next day, so we would get that data too late, if we always requested for example from 00.00 to 00.00
    fetchPeriodStartDate := time.Now().UTC()
    fetchPeriodStartDate = time.Date(
      fetchPeriodStartDate.Year(),
      fetchPeriodStartDate.Month(),
      fetchPeriodStartDate.Day(),
      22, //Hour
      00, //minute
      00, //second
      0,  //nanosecond
      time.UTC,
    )

    if(!haveToday){
      fetchPeriodStartDate = fetchPeriodStartDate.AddDate(0,0,-1)
    }
    err := getNewData(plan,fetchPeriodStartDate)
    if(err != nil){

      //if we were trying to fetch today's prices, and it didn't succeed, that is an error.
      //If we were trying to fetch tomorrow's prices and it didn't succeed, it's okay. We can try again next hour,
      //until the day changes.
      if(!haveToday){
        return err
      }else{
        fmt.Println("Tried but didn't get tomorrow's data: " + err.Error())
      }

    }else{
      jsonrw.WriteToJsonFile(filename,plan)
    }
  }
  return nil
}


func getNewData(plan *[]PlanData, fetchPeriodStartDate time.Time)error{
  client := http.Client{
    Timeout: 60 * time.Second,
  }
  var err error
  temperature, err := temperature.GetTemperature(client,fetchPeriodStartDate);
  if err != nil{
    return err
  }
  fmt.Println(temperature)
  endDate := fetchPeriodStartDate.AddDate(0,0,1)

  prices,err := electricityprice.GetPrices(fetchPeriodStartDate.Format("200601021504"), endDate.Format("200601021504"),client);
  if(err != nil){
    return err
  }
  var limits []int
  limits,err = getLimits(temperature, prices, plan, fetchPeriodStartDate)
  if(err != nil){
    return err
  }

  startTime := time.Date(
    fetchPeriodStartDate.Year(),
    fetchPeriodStartDate.Month(),
    fetchPeriodStartDate.Day(),
    22, //Hour
    00, //minute
    00, //second
    0,  //nanosecond
    time.UTC,
  )

  for i:= range limits{
    var newHour PlanData
    newHour.Limit = limits[i]
    newHour.Time = startTime.Format("020120061504")
    startTime = startTime.Add(time.Hour)
    *plan = append( *plan, newHour)
  }

  return nil
}




func FindLimitForNow(plan *[]PlanData)(int,error){
  curTime := time.Now().UTC()

  if(len(*plan) == 0){
    return 0, errors.New("Plan length 0")
  }

  index := len(*plan)-1
  latestPlannedHourTimeStr := (*plan)[index].Time
  latestPlannedHourTime,err := time.Parse("020120061504",latestPlannedHourTimeStr)
  if(err != nil){
    return 0,err
  }
  for latestPlannedHourTime.After(curTime){
    index--
    if(index < 0){
      return 0, errors.New("Current hour not found (index<0)")
    }
    latestPlannedHourTimeStr = (*plan)[index].Time
    latestPlannedHourTime,err = time.Parse("020120061504",latestPlannedHourTimeStr)
    if(err != nil){
      return 0,err
    }
  }
  fmt.Println((*plan)[index].Time)
  fmt.Println(curTime)
  return (*plan)[index].Limit,nil
}
