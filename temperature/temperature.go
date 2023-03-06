package temperature


import(
  "net/http"
  "io/ioutil"
  "encoding/json"
  "errors"
  "home-heating/secret"
  "time"
)


const REQUEST = "https://api.openweathermap.org/data/2.5/forecast?"+ secret.LOCATION +"&units=metric&appid="+secret.WEATHER_TOKEN

type TemperatureData struct{
  Count int `json:"cnt"`
  List []struct{
    Time int64 `json:"dt"`
    Main struct{
      Temperature float64 `json:"temp"`
    }`json:"main"`
  }`json:"list"`
}

//return (temp,errorcode)
//returns average temperature from the next 24h starting from timeNow (unix, UTC)
func GetTemperature(client http.Client, startTime time.Time)(float64, error){
  resp, err := client.Get(REQUEST)
  if err != nil {
    return 0,errors.New("Temperature fetch error1: " + err.Error())
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return 0,errors.New("Temperature fetch error2: " + err.Error());
  }


  var dat TemperatureData;
  if err := json.Unmarshal(body, &dat); err != nil {
      return 0,errors.New("Temperature fetch error3: " + err.Error())
  }

  if(resp.StatusCode != 200){
    return 0,errors.New("Temperature fetch error4: " +resp.Status);
  }
  var temperature float64 = 0.0;
  var count int = 0;
  for i:= range dat.List{

    //we only care about the next 24h
    if((dat.List[i].Time - startTime.Unix()) > 24*3600){
      break
    }

    if(dat.List[i].Time >= startTime.Unix()){
      temperature += dat.List[i].Main.Temperature;
      count++
    }
  }

  if(count == 0){
    return 0,errors.New("Didn't receive data for the date specified" + string(body))
  }

  temperature /= float64(count);
  return temperature,nil;
}
