package shelly

import(
  "net/http"
  "strconv"
  "errors"
)


const SWITCH_MAX_ON_TIME = 3600*3;





func SetSwitch(id uint8, value bool, client http.Client)error{
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
