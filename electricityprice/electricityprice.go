package electricityprice

import(
  "net/http"
  "io/ioutil"
  "encoding/xml"
  "errors"
  "time"
  "home-heating/secret"
)




type PriceData struct{
  PeriodStart string `xml:"period.timeInterval>start"`
  PeriodEnd string `xml:"period.timeInterval>end"`
  Prices []float64 `xml:"TimeSeries>Period>Point>price.amount"`
  Resolution string `xml:"TimeSeries>Period>resolution"`
  ErrorText string `xml:"Reason>text"`
}

func GetPrices(periodStart string, periodEnd string, client http.Client)([]float64, error){
  resp, err := client.Get(
    "https://web-api.tp.entsoe.eu/api?securityToken="+secret.PRICE_TOKEN+"&documentType=A44&in_Domain=10YFI-1--------U&out_Domain=10YFI-1--------U&periodStart="+periodStart+"&periodEnd="+periodEnd)

  if err != nil {
    return nil,err;
  }
  if(resp.StatusCode != 200){
    return nil,errors.New(resp.Status);
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil,err;
  }
  //fmt.Println(string(body))
  var dat PriceData
  if err := xml.Unmarshal(body, &dat); err != nil {
      return nil,err
  }
  if(dat.ErrorText != ""){
    return nil, errors.New(dat.ErrorText)
  }
  if(dat.Resolution != "PT60M"){
    return nil, errors.New("Resolution: "+dat.Resolution)
  }



  requestedPeriodStart,err := time.Parse("200601021504", periodStart)
  if(err != nil){
    return nil, err
  }
  receivedPeriodStart,err := time.Parse("2006-01-02T15:04Z",dat.PeriodStart)
  if(err != nil){
    return nil, err
  }

  requestedPeriodEnd,err := time.Parse("200601021504", periodEnd)
  if(err != nil){
    return nil, err
  }
  receivedPeriodEnd,err := time.Parse("2006-01-02T15:04Z",dat.PeriodEnd)
  if(err != nil){
    return nil, err
  }

  //the api gives data only in specific 24h chunks. So it might give more data than requested. Ignore the extra data
  if(receivedPeriodStart.After(requestedPeriodStart)){
    return nil, errors.New("Didn't receive requested data\nreceivedPeriodStart.After(requestedPeriodStart)\n" + string(body))
  }
  for(receivedPeriodStart.Before(requestedPeriodStart)){
    dat.Prices = dat.Prices[1:]
    receivedPeriodStart = receivedPeriodStart.Add(time.Hour)
  }

  if(receivedPeriodEnd.Before(requestedPeriodEnd)){
    return nil, errors.New("Didn't receive requested data\nreceivedPeriodEnd.Before(requestedPeriodEnd)\n"+ string(body))
  }
  for(receivedPeriodEnd.After(requestedPeriodEnd)){
    dat.Prices = dat.Prices[:len(dat.Prices)-1]
    receivedPeriodEnd = receivedPeriodEnd.Add(-time.Hour)
  }

  return dat.Prices,err;
}
