// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// #docker cp assignment.go cli:/opt/gopath/src/github.com/hyperledger/fabric/peer
// #docker exec -it cli bash

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================
// mkdir -p /opt/gopath/src/github.com/hyperledger/fabric/examples/chaincode/go/accounts
// cp ./assignment.go /opt/gopath/src/github.com/hyperledger/fabric/examples/chaincode/go/accounts/

// export CHANNEL_NAME=mychannel
// export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
// export PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

// export FABRIC_LOGGING_SPEC=info
// export CORE_PEER_LOCALMSPID="Org1MSP"
// export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
// export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
// export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

// ==== Install chiancode ====
// peer chaincode install -n accounts -v 1.0 -p github.com/hyperledger/fabric/examples/chaincode/go/accounts

// ==== Instantiate chiancode ====
// peer chaincode instantiate -o orderer.example.com:7050 --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n accounts -v 1.0 -c '{"Args":["init"]}'

// ==== Invoke transactions/ create accounts ====
// peer chaincode invoke -o orderer.example.com:7050  --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n accounts -c '{"Args":["createaccount", "A12345", "alice","200"]}'
// peer chaincode invoke -o orderer.example.com:7050  --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n accounts -c  '{"Args":["createaccount", "B12345", "bob","200"]}'
// peer chaincode invoke -o orderer.example.com:7050  --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n accounts -c '{"Args":["createaccount", "C12345", "charlie","200"]}'
// peer chaincode invoke -o orderer.example.com:7050  --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n accounts -c '{"Args":["createaccount", "D12345", "dave","200"]}'

// ==== Query accounts ====
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","alice"]}'
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","bob"]}'
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","charlie"]}'
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","dave"]}'
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","dale"]}'

// ==== Transfer amount ====
// peer chaincode invoke -o orderer.example.com:7050  --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n accounts -c '{"Args":["transfer","alice","bob", "50"]}'
// peer chaincode invoke -o orderer.example.com:7050  --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n accounts -c '{"Args":["transfer","dave","charlie", "50"]}'

// ==== Query accounts After transfer ====
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","alice"]}'
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","bob"]}'
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","charlie"]}'
// peer chaincode query -C $CHANNEL_NAME -n accounts -c '{"Args":["getaccount","dave"]}'

// Rich Query (Only supported if CouchDB is used as state database):
// peer chaincode query -C mychannel -n accounts -c '{"Args":["query","{\"selector\":{\"balance\": { \"$lt\":200}}}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// AccountChaincode example simple Chaincode implementation
type AccountChaincode struct {
}

type account struct {
	ObjectType string `json:"docType"`
	AccountID  string `json:"accountid"`
	Name       string `json:"name"`
	Balanace   int    `json:"balance"`
}

// Init initializes chaincode
// ===========================
func (t *AccountChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("====== Instantiating AccountChaincode ")
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *AccountChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createaccount" { //create a new account
		return t.createAccount(stub, args)
	} else if function == "getaccount" { //get account details
		return t.getAccount(stub, args)
	} else if function == "transfer" { //transfer given amount from one account to the other
		return t.transfer(stub, args)
	} else if function == "query" { //query all the accounts having balance < 'input'
		return t.queryAccountByBalance(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// createAccount - create a new account, store into chaincode state
// ============================================================
func (t *AccountChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1       2
	// "1234", "alice", "1000"
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// ==== Input sanitation ====
	fmt.Println("- start create account")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	accountID := args[0]
	name := strings.ToLower(args[1])
	balance, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if account already exists ====
	nameInBytes, err := stub.GetState(name)
	if err != nil {
		return shim.Error("Failed to get account: " + err.Error())
	} else if nameInBytes != nil {
		fmt.Println("This account already exists: " + name)
		return shim.Error("This account already exists: " + name)
	}

	// ==== Create account object and marshal to JSON ====
	objectType := "account"
	account := &account{objectType, accountID, name, balance}
	accountJSONasBytes, err := json.Marshal(account)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save account to state ===
	err = stub.PutState(name, accountJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end create account")
	return shim.Success(nil)
}

// ===============================================
// getAccount - read an account from chaincode state
// ===============================================
func (t *AccountChaincode) getAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the account to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the account from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Account does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ======================================
// transfer amount to a new account
// ======================================
func (t *AccountChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1       2
	// "owner", "bob", "amount"
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	accountName := args[0]
	transfereeName := strings.ToLower(args[1])
	balance, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}
	fmt.Println("- start transfer balance ", accountName, transfereeName)

	accountAsBytes, err := stub.GetState(accountName)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get account: %s", accountName) + err.Error())
	} else if accountAsBytes == nil {
		return shim.Error(fmt.Sprintf("Account: %s does not exist", accountName))
	}
	originalAcc := account{}
	err = json.Unmarshal(accountAsBytes, &originalAcc) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	//Deduct the amount from original account
	originalAcc.Balanace = originalAcc.Balanace - balance

	// update the original account
	accountJSONasBytes, _ := json.Marshal(originalAcc)
	err = stub.PutState(accountName, accountJSONasBytes) //rewrite the account
	if err != nil {
		return shim.Error(err.Error())
	}

	transfereeAsBytes, err := stub.GetState(transfereeName)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get account: %s", transfereeName) + err.Error())
	} else if accountAsBytes == nil {
		return shim.Error(fmt.Sprintf("Account: %s does not exist", transfereeName))
	}
	transfereeAcc := account{}
	err = json.Unmarshal(transfereeAsBytes, &transfereeAcc) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	//Add the amount to transferee account
	transfereeAcc.Balanace = transfereeAcc.Balanace + balance

	// update the transferee account
	transfereeJSONasBytes, _ := json.Marshal(transfereeAcc)
	err = stub.PutState(transfereeName, transfereeJSONasBytes) //rewrite the account
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("- end transfer balance (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

// =======Rich queries =========================================================================
// queryAccountByBalance uses a query string to perform a query for accounts.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *AccountChaincode) queryAccountByBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func main() {
	err := shim.Start(new(AccountChaincode))
	if err != nil {
		fmt.Printf("Error starting AccountChaincode: %s", err)
	}
}
