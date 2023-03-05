package main


import(
  "os"
  "home-heating/errorreport"
  "home-heating/config"
  "fmt"
)


func main(){
  dat, err := os.ReadFile(config.ERROR_FILE_NAME)
  if(err != nil){
    panic(err)
  }
  if(len(string(dat)) != 0){
    fmt.Println("Odottamaton error",string(dat))
    errorreport.Report("Odottamaton error",string(dat),config.ERROR_CODE_PANIC)
  }
  os.Truncate(config.ERROR_FILE_NAME,0)
}
