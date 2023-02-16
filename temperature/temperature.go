package temperature


import(
  "net/http"
  "io/ioutil"
  "encoding/json"
  "log"
  "errors"
)

const TOKEN = "WEATHER_TOKEN"

const REQUEST = "https://api.openweathermap.org/data/2.5/forecast?LOCATION&cnt=8&units=metric&appid="+TOKEN

type TemperatureData struct{
  Count int `json:"cnt"`
  List []struct{
    Main struct{
      Temperature float64 `json:"temp"`
    }`json:"main"`
  }`json:"list"`
}

//return (temp,errorcode)
func GetTemperature(client http.Client)(float64, error){
  defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()


  resp, err := client.Get(REQUEST)
  if err != nil {
    return 0,err;
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return 0,err;
  }


  var dat TemperatureData;
  if err := json.Unmarshal(body, &dat); err != nil {
      return 0,err
  }

  if(resp.StatusCode != 200){
    return 0,errors.New(resp.Status);
  }
  var temperature float64 = 0.0;
  for i:= range dat.List{
    temperature += dat.List[i].Main.Temperature;
  }
  temperature /= float64(dat.Count);
  return temperature,nil;
}
