package limitplan
import(
  "errors"
  "fmt"
  "sort"
  "time"
  "os"
  "io/ioutil"
  "encoding/json"
  "net/http"
  "home-heating/electricityprice"
  "home-heating/temperature"
)


type OneHourPlanData struct{
  Time string //format: ddmmyyyyhhmm
  Limit int
}

type PlanData struct{
  Hours []OneHourPlanData
}


const PASSIIVISET_TUNNIT_KUN_PAKKASTA_10 = 5
const MAKSIMI_TEHO = 10 //kWh
const MINIMI_TEMP = (24*MAKSIMI_TEHO-101.51)/(-4.59)  //Missä lämpötilassa lämmitys pitää olla koko ajan päällä. Perustuu energyUsage-function tietoihin.


const NO_LIMIT=3
//SMALL_LIMIT:=2
//BIG_LIMIT=1
const TOTAL_LIMIT=0

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
  var TEHO_KUN_PAKKASTA_10 float64= energyUsage(-10)/(24-PASSIIVISET_TUNNIT_KUN_PAKKASTA_10)

  //lineaarinen sovitus tavoitellulle teholle siten käyttäen arvoja (MINIMI_TEMP,MAKSIMITEHO), (-10,TEHO_KUN_PAKKASTA_10)
  return ((MAKSIMI_TEHO-TEHO_KUN_PAKKASTA_10)/(MINIMI_TEMP+10))*(temperature-MINIMI_TEMP)+MAKSIMI_TEHO;
}


func getLimits(temperature float64, prices []float64)([]int,error){
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
  var plan []int
  for i:=0;i<24;i++{
    plan = append(plan,NO_LIMIT)
  }

  var passiveTime = passiveTime(temperature)

  for i:=0; i<passiveTime; i++{
    plan[pricesWithIndex[i].Index] = TOTAL_LIMIT;
  }

  addBuffer(&plan, temperature)

  return plan,nil

}


//after total limit we will limit the following hours to prevent power surge after TOTAL_LIMIT
//the limit will last the same time that the total limit lasted
//half of the limit time there will be smaller limit than the planned power
//the the other half there will be larger limit than the planned power
func addBuffer(plan *[]int, temperature float64){
  var plannedPower = plannedPowerWhenActive(temperature)


  //Oletetaan että lämpötila liikkuu välillä -20...15
  //Valitaan kaksi tehorajoitusarvoa tältä väliltä tasaisesti
  powerLimits := []float64{0, plannedPowerWhenActive(3.34), plannedPowerWhenActive(-8.32), MAKSIMI_TEHO}
  //fmt.Println(powerLimits)


  //find a limit thats closest to the plannedpower (But the limit power is smaller than planned power)
  var closestPowerLimit = TOTAL_LIMIT
  for (powerLimits[closestPowerLimit+1] < plannedPower) && (closestPowerLimit < NO_LIMIT){
    closestPowerLimit++
  }
  //fmt.Println("Planned: ", plannedPower, "Limit: ", powerLimits[closestPowerLimit])


  var totalLimitHoursSequental = 0

  for i:= range *plan{
    if((*plan)[i] == TOTAL_LIMIT){
      totalLimitHoursSequental++
    }else{
      if(totalLimitHoursSequental > 0){
        limitTime := 0

        //how many not limited hours there are after total limit
        //we will limit at max the same time that the total limit lasted
        for j:=0; j<totalLimitHoursSequental && j+i<len(*plan) && (*plan)[j+i] == NO_LIMIT; j++{
          limitTime++
        }


        for j:=0; j<limitTime; j++{
          if(j < limitTime/2){
            if(closestPowerLimit == TOTAL_LIMIT){
              (*plan)[j+i] = closestPowerLimit+1
            }else{
              (*plan)[j+i] = closestPowerLimit
            }
          }else{
            (*plan)[j+i] = closestPowerLimit+1
          }
        }

        i+=limitTime
      }

      totalLimitHoursSequental = 0
    }
  }
}


func UpdatePlan(filename string, plan *PlanData)error{
  //we will be fetching tomorrow's prices
  nextDateString := time.Now().UTC().AddDate(0,0,1).Format("02012006")

  //check if we already have tomorrow's plan
  var lastPlanHourDate string = ""
  if(len(plan.Hours) != 0){
    lastPlanHourDate = plan.Hours[len(plan.Hours)-1].Time

    //remove 4 last characters (clock information)
    lastPlanHourDate = lastPlanHourDate[0:len(lastPlanHourDate)-4]
  }
  if(lastPlanHourDate != nextDateString){

    //even though we will be fetching tomorrow's data, the fetch period starts today. This is because the price API considers days to begin at 2300 (or 2200 during summertime)
    fetchPeriodStartDate := time.Now().UTC()

    //before adding a plan for tomorrow, check if we have today's plan (if not, an error has ocurred)
    curDateString := time.Now().UTC().Format("02012006")
    if(lastPlanHourDate != curDateString){
      fetchPeriodStartDate = fetchPeriodStartDate.AddDate(0,0,-1)
    }

    err := getNewData(plan,fetchPeriodStartDate)
    if(err != nil){

      //if we were trying to fetch today's prices, and it didn't succeed, that is an error.
      //If we were trying to fetch tomorrow's prices and it didn't succeed, it's okay. We can try again next hour,
      //until the day changes.
      if(fetchPeriodStartDate.Format("02.01.2006") == time.Now().UTC().AddDate(0,0,-1).Format("02.01.2006")){
        return err
      }

    }else{
      WritePlansToJson(filename,plan)
    }
  }
  return nil
}


func getNewData(plan *PlanData, fetchPeriodStartDate time.Time)error{
  client := http.Client{
    Timeout: 60 * time.Second,
  }
  var err error
  temperature, err := temperature.GetTemperature(client);
  if err != nil{
    return err
  }

  startDateString := fetchPeriodStartDate.Format("20060102")
  endDate := fetchPeriodStartDate.AddDate(0,0,1)
  endDateString := endDate.Format("20060102")


  //we request tomorrows date beginning from today 2200 and ending tomorrow 2200
  //This is because our API considers days to start (in Finland) at 2300 or 2200 depending on weather we live in summer time or winter time
  //The last hours of a day are published with the next day, so we would get that data too late, if we always requested for example from 00.00 to 00.00
  prices,err := electricityprice.GetPrices(startDateString+"2200", endDateString+"2200",client);
  if(err != nil){
    return err
  }
  var limits []int
  limits,err = getLimits(temperature, prices)
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
    var newHour OneHourPlanData
    newHour.Limit = limits[i]
    newHour.Time = startTime.Format("020120061504")
    startTime = startTime.Add(time.Hour)
    (*plan).Hours = append( (*plan).Hours, newHour)
  }

  return nil
}



func FindLimitForNow(plan *PlanData)(int,error){
  curTime := time.Now().UTC()

  if(len((*plan).Hours) == 0){
    return 0, errors.New("Plan length 0")
  }

  index := len((*plan).Hours)-1
  latestPlannedHourTimeStr := (*plan).Hours[index].Time
  latestPlannedHourTime,err := time.Parse("020120061504",latestPlannedHourTimeStr)
  if(err != nil){
    return 0,err
  }
  for latestPlannedHourTime.Sub(curTime) > time.Hour{
    index--
    if(index < 0){
      return 0, errors.New("Current hour not found (index<0)")
    }
    latestPlannedHourTimeStr = (*plan).Hours[index].Time
    latestPlannedHourTime,err = time.Parse("020120061504",latestPlannedHourTimeStr)
    if(err != nil){
      return 0,err
    }
  }

  return (*plan).Hours[index].Limit,nil
}


func GetPlansFromJson(filename string, plan *PlanData)error{
  _,err := os.Stat(filename)
  var jsonFile *os.File
  if(err != nil){
    jsonFile,err = os.Create(filename)
  }else{
    jsonFile, err = os.Open(filename)
  }
  defer jsonFile.Close()
  if err != nil {
    return err
  }


  byteVal,err := ioutil.ReadAll(jsonFile)
  if(err != nil){
    return err
  }
  err = json.Unmarshal(byteVal, plan)

  return err
}



func WritePlansToJson(filename string, plan *PlanData)error{
  jsonFile,err := os.Create(filename)
  defer jsonFile.Close()

  toWrite,err := json.Marshal(plan)
  if(err != nil){
    return err
  }

  _,err = jsonFile.Write(toWrite)
  fmt.Println(string(toWrite))
  return err

}
