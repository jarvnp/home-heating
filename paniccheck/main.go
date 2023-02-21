package main


import(
  "os"
  "home-heating/errorreport"
  "fmt"
)


func main(){
  dat, err := os.ReadFile("/home/matias/go/src/home-heating/error")
  if(err != nil){
    panic(err)
  }
  if(len(string(dat)) != 0){
    fmt.Println("Odottamaton error",string(dat))
    errorreport.Report("Odottamaton error",string(dat))
  }
  os.Truncate("/home/matias/go/src/home-heating/error",0)
}
