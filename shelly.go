package main

import(
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "log"
  )

//return (temp,errorcode)
func getTemperature()(float64, error){
  defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()


  resp, err := http.Get("https://api.openweathermap.org/data/2.5/forecast?LOCATION&cnt=8&units=metric&appid=WEATHER_TOKEN")
  if err != nil {
    return 0,err;
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return 0,err;
  }
  var dat map[string]interface{}
  if err := json.Unmarshal(body, &dat); err != nil {
      return 0,err
  }

  if(dat["cod"].(string) != "200"){
    return 0,dat["cod"].(error);
  }
  var cnt float64 = dat["cnt"].(float64);

  list := dat["list"].([]interface{})

  var temperature float64 = 0.0;
  for i:= range list{
    listItem := list[i].(map[string]interface{})
    main := listItem["min"].(map[string]interface{})
    temperature += main["temp"].(float64);
  }

  temperature /= cnt;
  return temperature,nil;
}


func main() {


  resp, err:= getTemperature();
  if(err == nil){
    fmt.Println(resp);
  }




}
