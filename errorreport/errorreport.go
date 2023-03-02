package errorreport

import(
  "home-heating/email"
  "net/http"
  "time"
  "fmt"
  "home-heating/config"
  "home-heating/secret"
  "home-heating/jsonrw"
)

const HEADER = "Lämmityssysteemin ERROR!"

const RECOVERY_HEADER = "Lämmityssysteemi toimii taas"
const RECOVERY_MESSAGE = "Lämmityssysteemi on palannut normaaliin toimintaan"


type errorData struct{
  ErrorCode int
  Time string //ddmmyyyyhhmm
}



func timeStrToTime(timeStr string)(time.Time,error){
  time,err := time.Parse("020120061504", timeStr)
  return time,err
}



//checks if we have recenlty reported on the same error
func isRecentlyReported(errorCode int)(bool, error){
  var errorHistory []errorData
  err := jsonrw.ReadFromJsonFile("errorhistory.json",&errorHistory)
  if(err != nil){
    return false,err
  }
  timeNow := time.Now().UTC()
  for i:= range errorHistory{
    if(errorHistory[i].ErrorCode == errorCode){
      errorTime,err := timeStrToTime(errorHistory[i].Time)
      if(err != nil){
        return false,err
      }
      if(timeNow.Sub(errorTime) < time.Hour*config.DELAY_BETWEEN_ERROR_MESSAGES){
        return true,nil
      }
    }
  }
  return false, nil
}


//stores info about made error report
func storeError(errorCode int)(error){
  var errorHistory []errorData

  err:= jsonrw.ReadFromJsonFile("errorhistory.json", &errorHistory)
  if(err != nil){
    return err
  }
  var newError errorData
  newError.Time = time.Now().UTC().Format("020120061504")
  newError.ErrorCode = errorCode

  //remove old errors with same errorcode
  for i:= range errorHistory{
    if(errorHistory[i].ErrorCode == errorCode){
      //replace the to-be-removed element with the last element, and remove last element
      errorHistory[i] = errorHistory[len(errorHistory)-1]
      errorHistory = errorHistory[:len(errorHistory)-1]
    }
  }

  errorHistory = append(errorHistory,newError)

  err = jsonrw.WriteToJsonFile("errorhistory.json",&errorHistory)

  return err
}


func Report(description string, errorText string, errorCode int){
  client := http.Client{
    Timeout: 60 * time.Second,
  }

  //while testing:
  fmt.Println("error: ", description + "\n\n" + errorText,client)

  var isReported bool = false
  var err error
  //always report if there has been panic
  if(errorCode != config.ERROR_CODE_PANIC){
    isReported, err = isRecentlyReported(errorCode)
    if(err != nil){
      panic(err)
    }
  }

  if(isReported && (errorCode != config.ERROR_CODE_PANIC)){
    fmt.Println("No need to report")
  }else{
    storeError(errorCode)
    if(config.ENABLE_EMAIL_REPORTS){
      err:= email.SendEmail(secret.RECIPIENTS,HEADER,description + "\n\n" + errorText,client)
      if(err != nil){
        panic("Emailin lähetys ei onnistu:\n\n"+err.Error())
      }
    }
  }

}

//returns true if there has previously been an error
func IsRecovery()(bool){
  var errorHistory []errorData
  err := jsonrw.ReadFromJsonFile("errorhistory.json",&errorHistory)
  if(err != nil){
    panic(err)
  }
  if(len(errorHistory) != 0){
    return true
  }
  return false
}

func ClearErrorHistory(){
  var errorHistory []errorData
  err := jsonrw.WriteToJsonFile("errorhistory.json",&errorHistory)
  if(err != nil){
    panic(err)
  }
}



func ReportRecovery(){


  //while debugging:
  fmt.Println(RECOVERY_HEADER + "\n\n"  + RECOVERY_MESSAGE)

  if(config.ENABLE_EMAIL_REPORTS){
    client := http.Client{
      Timeout: 60 * time.Second,
    }
    err:= email.SendEmail(secret.RECIPIENTS,RECOVERY_HEADER,RECOVERY_MESSAGE,client)
    if(err != nil){
      panic("Emailin lähetys ei onnistu:\n\n"+err.Error())
    }
  }
}
