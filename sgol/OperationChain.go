package sgol

import (
  "fmt"
  "encoding/json"
  "crypto/md5"
)

type OperationChain struct {
  Name string `json:"name" yaml:"name" hcl:"name"`
  Operations []Operation `json:"operations" yaml:"operations" hcl:"operations"`
  OutputType string `json:"output_type" "yaml":"output_type" hcl:"output_type"`
  Limit int `json:"limit" bson:"limit" yaml:"limit" hcl:"limit"`
}

func (oc *OperationChain) Hash() (string, error) {
  data, err := json.Marshal(oc)
  if err != nil {
    return "", err
  }
  hash := fmt.Sprintf("%x", md5.Sum(data))
  return hash, nil
}


func NewOperationChain(name string, operations []Operation, outputType string) OperationChain {

  chain := OperationChain{
    Name: name,
    OutputType: outputType,
  }

  if len(operations) > 0 {
    chain.Operations = operations
    fmt.Println("Op TYPE:", operations[len(operations) - 1].GetTypeName())
    if operations[len(operations) - 1].GetTypeName() == "LIMIT" {
      chain.Limit = operations[len(operations) - 1].(OperationLimit).Limit
    }
  }

  return chain

}
