package shelly

import (
	"errors"
	"home-heating/config"
	"home-heating/secret"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const REQUEST_WAIT_DELAY = time.Second*2;

func SetLimit(limit int)error{
  client := http.Client{
    Timeout: 120 * time.Second,
  }
  err:=setSwitch(0,config.SHELLY_OUTPUT_STATES_FOR_LIMITS[limit][0],client)
  if(err != nil){
    return err
  }

  err=setSwitch(1,config.SHELLY_OUTPUT_STATES_FOR_LIMITS[limit][1],client)
  return err
}



func setSwitch(id uint8, value bool, client http.Client)error{
  data := url.Values{}
  data.Set("id", secret.SHELLY_ID)
  data.Set("auth_key", secret.CLOUD_TOKEN)
  data.Set("method", "Switch.Set")
  var onString string;
  var toggleString string = "";
  if value == true{
    onString = "true"
    toggleString = `,"toggle_after":` + strconv.Itoa(int(config.SHELLY_SWITCH_MAX_ON_TIME));
  }else{
    onString = "false"
  }
  data.Set("params", `{"id":` +strconv.Itoa(int(id))+ `, "on": ` + onString + toggleString + `}`)

  resp, err := client.PostForm(secret.CLOUD_SERVER,data)
  if err != nil {
    return err;
  }
  if(resp.StatusCode != 200){
    bodyBytes, _ := io.ReadAll(resp.Body)
    return errors.New(string(bodyBytes));
  }

  time.Sleep(REQUEST_WAIT_DELAY);  // There is a restiction on API use
  err = resetShellyWatchdogScript(client)
  return err;
}

//stop and start again a watchdog script.
//the script will send warning email, if it is not reset periodically
func resetShellyWatchdogScript(client http.Client)error{

  data := url.Values{}
  data.Set("id", secret.SHELLY_ID)
  data.Set("auth_key", secret.CLOUD_TOKEN)
  data.Set("method", "Script.Stop")
  data.Set("params", `{"id":` +config.SHELLY_WATCHDOG_SCRIPT_ID + `}`)
  resp, err := client.PostForm(secret.CLOUD_SERVER,data)
  if err != nil {
    return err;
  }
  if(resp.StatusCode != 200){
    bodyBytes, _ := io.ReadAll(resp.Body)
    return errors.New(string(bodyBytes));
  }
  time.Sleep(REQUEST_WAIT_DELAY);  // There is a restiction on API use

  data.Set("method", "Script.Start")

  resp, err = client.PostForm(secret.CLOUD_SERVER,data)
  if err != nil {
    return err;
  }
  if(resp.StatusCode != 200){
    bodyBytes, _ := io.ReadAll(resp.Body)
    return errors.New(string(bodyBytes));
  }

  time.Sleep(REQUEST_WAIT_DELAY);  // There is a restiction on API use
  return nil
}
