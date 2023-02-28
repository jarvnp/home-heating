package main


import(
  "os"
  "home-heating/errorreport"
  "fmt"
)


func main(){
  dat, err := os.ReadFile("error")
  if(err != nil){
    panic(err)
  }
  if(len(string(dat)) != 0){
    fmt.Println("Odottamaton error",string(dat))
    errorreport.Report("Odottamaton error",string(dat))
  }
  os.Truncate("error",0)
}
