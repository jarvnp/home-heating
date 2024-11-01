package electricityprice

import (
	"encoding/xml"
	"errors"
	"home-heating/secret"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)



type Point struct{
  Price float64 `xml:"price.amount"`
  Position int `xml:"position"`
}


type PriceData struct{
  PeriodStart string `xml:"period.timeInterval>start"`
  PeriodEnd string `xml:"period.timeInterval>end"`
  TimeSeries []struct{
    Points []Point`xml:"Period>Point"`
    TimeStart string `xml:"Period>timeInterval>start"`
    Resolution string `xml:"Period>resolution"`
  } `xml:"TimeSeries"`

  ErrorText string `xml:"Reason>text"`
}


// ChatGPT:
// FillMissingPoints takes an array of Points and fills in any missing positions (1 through 24)
// with the last known price.
func FillMissingPoints(points []Point) []Point {
	// First, sort the points array by Position to ensure order
	// Sorting can be omitted if the array is guaranteed to be in order
	sort.Slice(points, func(i, j int) bool {
		return points[i].Position < points[j].Position
	})

	// Create a map to hold the Position to Price mappings for quick lookups
	positionToPrice := make(map[int]float64)
	for _, point := range points {
		positionToPrice[point.Position] = point.Price
	}

	// FilledPoints will hold the result with no missing positions
	filledPoints := make([]Point, 0, 24)

	// Start with the price of the first known position
	lastKnownPrice := 0.0

	// Fill in positions from 1 to 24
	for i := 1; i <= 24; i++ {
		if price, exists := positionToPrice[i]; exists {
			// If the position exists, use the given price
			lastKnownPrice = price
		}
		// Add the point with the current position and last known price
		filledPoints = append(filledPoints, Point{
			Price:    lastKnownPrice,
			Position: i,
		})
	}

	return filledPoints
}

func GetPrices(periodStart string, periodEnd string, client http.Client)([]float64, error){
  resp, err := client.Get(
    "https://web-api.tp.entsoe.eu/api?securityToken="+secret.PRICE_TOKEN+"&documentType=A44&in_Domain=10YFI-1--------U&out_Domain=10YFI-1--------U&periodStart="+periodStart+"&periodEnd="+periodEnd)

  if err != nil {
    return nil,errors.New("Price fetch error1: " + err.Error());
  }
  if(resp.StatusCode != 200){
    return nil,errors.New("Price fetch error2: " + resp.Status);
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil,errors.New("Price fetch error3: " + err.Error());
  }
  //fmt.Println(string(body))
  var dat PriceData
  if err := xml.Unmarshal(body, &dat); err != nil {
      return nil,errors.New("Price fetch error4: " + err.Error());
  }
  if(dat.ErrorText != ""){
    return nil, errors.New("Price fetch error5: " + dat.ErrorText)
  }
  for _,v := range dat.TimeSeries{
    res := v.Resolution
    if(res != "PT60M"){
      return nil, errors.New("Resolution: "+res)
    }
  }

  var timeParseError error = nil;
  sort.Slice(dat.TimeSeries, func(i,j int)bool{
    timeA,err := time.Parse("2006-01-02T15:04Z",dat.TimeSeries[i].TimeStart)
    if(timeParseError == nil){
      timeParseError = err
    }
    timeB,err := time.Parse("2006-01-02T15:04Z",dat.TimeSeries[j].TimeStart)
    if(timeParseError == nil){
      timeParseError = err
    }
    return timeA.Before(timeB)
  })
  if timeParseError != nil {
    return nil, errors.New("Error parsing times. \n" + timeParseError.Error())
  }

  for i,series := range dat.TimeSeries{
    dat.TimeSeries[i].Points = FillMissingPoints(series.Points)
  }

  requestedPeriodStart,err := time.Parse("200601021504", periodStart)
  if(err != nil){
    return nil, errors.New("Price fetch error6: " + err.Error());
  }
  receivedPeriodStart,err := time.Parse("2006-01-02T15:04Z",dat.PeriodStart)
  if(err != nil){
    return nil, errors.New("Price fetch error7: " + err.Error());
  }

  requestedPeriodEnd,err := time.Parse("200601021504", periodEnd)
  if(err != nil){
    return nil, errors.New("Price fetch error8: " + err.Error());
  }
  receivedPeriodEnd,err := time.Parse("2006-01-02T15:04Z",dat.PeriodEnd)
  if(err != nil){
    return nil, errors.New("Price fetch error9: " + err.Error());
  }


  var allPrices []float64
  for _, series := range dat.TimeSeries {
    for _,point := range series.Points{
      allPrices = append(allPrices, point.Price)
    }
  }

  //the api gives data only in specific 24h chunks. So it might give more data than requested. Ignore the extra data
  if(receivedPeriodStart.After(requestedPeriodStart)){
    return nil, errors.New("Didn't receive requested data\nreceivedPeriodStart.After(requestedPeriodStart)\n" + string(body))
  }
  for(receivedPeriodStart.Before(requestedPeriodStart)){
    allPrices = allPrices[1:]
    receivedPeriodStart = receivedPeriodStart.Add(time.Hour)
  }

  if(receivedPeriodEnd.Before(requestedPeriodEnd)){
    return nil, errors.New("Didn't receive requested data\nreceivedPeriodEnd.Before(requestedPeriodEnd)\n"+ string(body))
  }
  for(receivedPeriodEnd.After(requestedPeriodEnd)){
    allPrices = allPrices[:len(allPrices)-1]
    receivedPeriodEnd = receivedPeriodEnd.Add(-time.Hour)
  }
  return allPrices,err;
}
