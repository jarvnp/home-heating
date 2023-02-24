package dayplan
import(
  "errors"
  //"fmt"
  "sort"
)




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


func Plan(temperature float64, prices []float64)([]int,error){

  //the linear models don't apply to too high temperatures
  if(temperature > 15.0){
    temperature = 15.0
  }


  //this *should* work even with daylight savings time, when there sometimes are 23 or 25 hours in a day.
  //This is because the prices are fetched using UTC time, which is unaffected by daylight savings
  if(len(prices) != 24){
    return nil, errors.New("Not enough price data")
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
