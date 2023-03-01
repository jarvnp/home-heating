package jsonrw

import(
  "encoding/json"
  "os"
  "io/ioutil"
)


func ReadFromJsonFile(filename string, data interface{})error{
  jsonFile,err:= os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

  defer jsonFile.Close()
  if err != nil {
    return err
  }


  byteVal,err := ioutil.ReadAll(jsonFile)
  if(len(byteVal) == 0){
    _,err = jsonFile.WriteString("null")  //empty json
    if(err != nil){
      return err
    }
    byteVal = append(byteVal,[]byte("null")...)
  }
  if(err != nil){
    return err
  }
  err = json.Unmarshal(byteVal, data)

  return err
}



func WriteToJsonFile(filename string, data interface{})error{
  jsonFile,err := os.Create(filename)
  defer jsonFile.Close()

  toWrite,err := json.Marshal(data)
  if(err != nil){
    return err
  }

  _,err = jsonFile.Write(toWrite)
  return err

}
