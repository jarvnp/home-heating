package main

import(
  "home-heating/dayplan"
  "bufio"
  "fmt"
  "os"
  "strconv"
)


func main(){
  readFile, err := os.Open("config/prices2022")

  if err != nil {
      fmt.Println(err)
  }
  fileScanner := bufio.NewScanner(readFile)

  fileScanner.Split(bufio.ScanLines)

  var prices [][]float64

  index := 0
  for fileScanner.Scan() {
      if(index %24 == 0){
        prices = append(prices,[]float64{})
      }
      price,_ := strconv.ParseFloat(fileScanner.Text(),64)
      prices[index/24] = append(prices[index/24], price)
      index++
  }
  //fmt.Println(prices)
  readFile.Close()



  readFile, err = os.Open("config/temps2022")

  if err != nil {
      fmt.Println(err)
  }
  fileScanner = bufio.NewScanner(readFile)

  fileScanner.Split(bufio.ScanLines)

  var temps []float64

  index = 0
  sum := 0.0
  for fileScanner.Scan() {
      if(index %24 == 0){
        temps = append(temps,sum/24.0)
        sum = 0
      }
      temp,_ := strconv.ParseFloat(fileScanner.Text(),64)
      sum += temp
      index++
  }
  //fmt.Println(temps)
  readFile.Close()


  for i:= range temps{
    plan, _ := dayplan.Plan(temps[i],prices[i]);
    for j:= range plan{
      fmt.Println(plan[j],prices[i][j])
    }
  }



}
