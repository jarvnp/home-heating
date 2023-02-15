package main

import(
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "log"
  "strconv"
  "errors"
  "time"
  )


const SWITCH_MAX_ON_TIME = 3600*3;


//return (temp,errorcode)
func getTemperature(client http.Client)(float64, error){
  defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()


  resp, err := client.Get("https://api.openweathermap.org/data/2.5/forecast?LOCATION&cnt=8&units=metric&appid=WEATHER_TOKEN")
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
    main := listItem["main"].(map[string]interface{})
    temperature += main["temp"].(float64);
  }

  temperature /= cnt;
  return temperature,nil;
}


func setSwitch(id uint8, value bool, client http.Client)error{
  var request string;
  if value == true{
    request = "SHELLY_ADDRESS/rpc/Switch.Set?id="+strconv.Itoa(int(id))+"&on=true&toggle_after="+strconv.Itoa(int(SWITCH_MAX_ON_TIME))
  }else{
    request = "SHELLY_ADDRESS/rpc/Switch.Set?id=" +strconv.Itoa(int(id))+"&on=false"
  }

  resp, err := client.Get(request)
  if err != nil {
    return err;
  }
  if(resp.StatusCode != 200){
    return errors.New(resp.Status);
  }
  return nil;
}



func main() {
  client := http.Client{
    Timeout: 10 * time.Second,
  }

  resp, err:= getTemperature(client);
  if(err == nil){
    fmt.Println(resp);
  }

  err=setSwitch(1,false, client);
  if(err != nil){
    fmt.Println(err);
  }
}
