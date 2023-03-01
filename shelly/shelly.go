package shelly

import(
  "net/http"
  "strconv"
  "errors"
  "home-heating/config"
  "home-heating/secret"
  "time"
)






func SetLimit(limit int)error{
  client := http.Client{
    Timeout: 30 * time.Second,
  }
  err:=setSwitch(0,config.SHELLY_OUTPUT_STATES_FOR_LIMITS[limit][0],client)
  if(err != nil){
    return err
  }

  err=setSwitch(1,config.SHELLY_OUTPUT_STATES_FOR_LIMITS[limit][1],client)
  return err
}



func setSwitch(id uint8, value bool, client http.Client)error{
  var request string;
  if value == true{
    request = secret.SHELLY_ADDRESS+"/rpc/Switch.Set?id="+strconv.Itoa(int(id))+"&on=true&toggle_after="+strconv.Itoa(int(config.SHELLY_SWITCH_MAX_ON_TIME))
  }else{
    request = secret.SHELLY_ADDRESS+"/rpc/Switch.Set?id=" +strconv.Itoa(int(id))+"&on=false"
  }

  resp, err := client.Get(request)
  if err != nil {
    return err;
  }
  if(resp.StatusCode != 200){
    return errors.New(resp.Status);
  }

  err = resetShellyWatchdogScript(client)
  return err;
}

//stop and start again a watchdog script.
//the script will send warning email, if it is not reset periodically
func resetShellyWatchdogScript(client http.Client)error{
  var request string
  request = secret.SHELLY_ADDRESS + "/rpc/Script.Stop?id=" + config.SHELLY_WATCHDOG_SCRIPT_ID
  resp, err := client.Get(request)
  if err != nil {
    return err;
  }
  if(resp.StatusCode != 200){
    return errors.New(resp.Status);
  }

  request = secret.SHELLY_ADDRESS + "/rpc/Script.Start?id=" + config.SHELLY_WATCHDOG_SCRIPT_ID
  resp, err = client.Get(request)
  if err != nil {
    return err;
  }
  if(resp.StatusCode != 200){
    return errors.New(resp.Status);
  }

  return nil
}
