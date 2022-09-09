// MIT License
//
// Copyright (c) 2020 Dmitrii Ustiugov and EASE lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bufio"
	//"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	//"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/ease-lab/vhive/examples/endpoint"
  
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/lambda"
)

const TimeseriesDBAddr = "10.96.0.84:90"

var (
	completed   int64
	latSlice    LatencySlice
	//portFlag    *int
	//grpcTimeout time.Duration
	//withTracing *bool
	//workflowIDs map[*endpoint.Endpoint]string
)


type DummyPayloadStruct struct {
  //"exanple: For the function chameleon - an example payload is { "num_of_rows": 1, "num_of_cols": 3  }
  // IMPORTANT -= Note the CamelCase naming inside the structure
  NumOfRows  int `json:"num_of_rows,omitempty"` 
  NumOfCols  int `json:"num_of_cols,omitempty"` 
}

func main() {
	endpointsFile := flag.String("endpointsFile", "endpoints.json", "File with endpoints' metadata")
	rps := flag.Float64("rps", 1.0, "Target requests per second")
	runDuration := flag.Int("time", 5, "Run the experiment for X seconds")
	latencyOutputFile := flag.String("latf", "lat.csv", "CSV file for the latency measurements in microseconds")
	//portFlag = flag.Int("port", 80, "The port that functions listen to")
	debug := flag.Bool("dbg", false, "Enable debug logging")

	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
		FullTimestamp:   true,
	})
	log.SetOutput(os.Stdout)
	if *debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging is enabled")
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Info("Reading the endpoints from the file: ", *endpointsFile)

	endpoints, err := readEndpoints(*endpointsFile)
	if err != nil {
		log.Fatal("Failed to read the endpoints file: ", err)
	}
  log.Infof("Read %v endpoints from the file", strconv.Itoa(len(endpoints)))

	/*
  workflowIDs = make(map[*endpoint.Endpoint]string)
	for _, ep := range endpoints {
		workflowIDs[ep] = uuid.New().String()
	}
  */

	realRPS := runExperiment(endpoints, *runDuration, *rps)

	writeLatencies(realRPS, *latencyOutputFile)
}

func readEndpoints(path string) (endpoints []*endpoint.Endpoint, _ error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &endpoints); err != nil {
		return nil, err
	}
	return
}

func runExperiment(endpoints []*endpoint.Endpoint, runDuration int, targetRPS float64) (realRPS float64) {
	var issued int

	//Start(TimeseriesDBAddr, endpoints, workflowIDs)
  var wg sync.WaitGroup

	timeout := time.After(time.Duration(runDuration) * time.Second)
	tick := time.Tick(time.Duration(1000/targetRPS) * time.Millisecond)
	start := time.Now()
loop:
	for {
		ep := endpoints[issued%len(endpoints)]
    wg.Add(1)
    go invokeServingFunction(ep, &wg)
    /*
		if ep.Eventing {
			go invokeEventingFunction(ep)
		} else {
			go invokeServingFunction(ep)
		}
    */
		issued++

		select {
		case <-timeout:
			break loop
		case <-tick:
			continue
		}
	} // end for

  log.Infof("Waiting for workers to finish")
	wg.Wait()
	log.Infof("Workers Completed!")

	duration := time.Since(start).Seconds()
	realRPS = float64(completed) / duration
	//addDurations(End())
	log.Infof("Issued / completed requests: %d, %d", issued, completed)
	log.Infof("Real / target RPS: %.2f / %v", realRPS, targetRPS)
	log.Println("Experiment finished!")
	return
}



func InvokeLambda(address string) (float64){

  type ResponseLatency struct {
    Latency float64 `json:"latency"`
  }
  var retval float64 = -1.0

  //-----------------------------------------------------------------------------------------
  // Create Lambda service client
  //-----------------------------------------------------------------------------------------
  sess := session.Must(session.NewSessionWithOptions(session.Options{
      SharedConfigState: session.SharedConfigEnable,
  }))
	log.Debug("After creating sess")

  client := lambda.New(sess, &aws.Config{Region: aws.String("us-east-1")})
	log.Debug("After creating client")

  // Set name of lambda function to be invoked
  //lambda_function_name := "chameleon"
  lambda_function_name := address

  // Get the 10 most recent items
  //request := getItemsRequest{"time", "descending", 10}
  //request := payload_struct{1,3} //"exanple: a payload of { "num_of_rows": 1, "num_of_cols": 3  }
  request := DummyPayloadStruct{1,3} //Using the same payload struct for all (three) functions
	log.Debug("After creating request")

  payload, err := json.Marshal(request)
	//log.Debug("After payload, err...")

  if err != nil {
      log.Warnf("Error marshalling " + lambda_function_name + " request")
      log.Warnf(err.Error())
      return retval
      //os.Exit(0)
  }

  log.Debug("After creating payload without error")
  result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String(lambda_function_name), Payload: payload})
  //log.Debug("After client.Invoke, err...")
  if err != nil {
      log.Warnf("Failed to invoke %v, err=%v", lambda_function_name, err.Error())
      return retval
      //os.Exit(0)
  }
  log.Debug("After client.Invoke - without error")
  //-----------------------------------------------------------------------------------------
  // Process Lambda respose
  //-----------------------------------------------------------------------------------------
  /* Sample returned object:
  {
      ExecutedVersion: "$LATEST",
      Payload: <sensitive>,
      StatusCode: 200
  }*/
  // If the status code is NOT 200, the call failed
  var ret1 int
  ret1 = int(*result.StatusCode)
  if ret1 != 200 {
    log.Warnf("Error returned by function " + lambda_function_name + ", StatusCode: " + strconv.Itoa(ret1))
    return retval
    //os.Exit(0)
  }
  log.Debug("SUCCESS!...resp.StatusCode is 200!")

  // First unmarshall Base64 encoded string to a regular string
  var resp string
  err = json.Unmarshal([]byte(result.Payload), &resp)
  //log.Debug("After unmarshall result")
  if err != nil {
    log.Warnf("Error unmarshalling to string: " + lambda_function_name + " resp")
    log.Warnf(err.Error())
    return retval
    //os.Exit(0)
  }
  log.Debug("After unmarshall result without error - string is:")
  log.Debug("Response structure is:")
  log.Debug(resp)

  // Second unmarshall regular string that has Json data into a Go struct
  // Make sure that the Go struct has members named in CamelCase
  var resp_struct ResponseLatency
  err = json.Unmarshal([]byte(resp), &resp_struct)
  if err != nil {
      log.Warnf("Error unmarshalling to structure: " + lambda_function_name + " resp_struct")
      log.Warnf(err.Error())
      return retval
      //os.Exit(0)
  }
  atomic.AddInt64(&completed, 1)
  log.Debug("After unmarshall resp without error")
  log.Debug("Response structure is: \n", resp_struct.Latency)
  log.Debug(resp_struct)
  retval = resp_struct.Latency
  return retval
} // end of InvokeLambda



/*  
func SayHello(address, workflowID string) {
	dialOptions := []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}
	if *withTracing {
		dialOptions = append(dialOptions, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	}
	conn, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()

	_, err = c.SayHello(ctx, &HelloRequest{
		Name: "faas",
		VHiveMetadata: vhivemetadata.MakeVHiveMetadata(
			workflowID,
			uuid.New().String(),
			time.Now().UTC(),
		),
	})
	if err != nil {
		log.Warnf("Failed to invoke %v, err=%v", address, err)
	} else {
		atomic.AddInt64(&completed, 1)
	}
}
*/

func invokeServingFunction(endpoint *endpoint.Endpoint, wg *sync.WaitGroup) {
	//defer getDuration(startMeasurement(endpoint.Hostname)) // measure entire invocation time
  defer wg.Done()

	//address := fmt.Sprintf("%s:%d", endpoint.Hostname, *portFlag)
	address := fmt.Sprintf("%s", endpoint.Hostname)
	log.Debug("Invoking: ", address)
  
  var retval float64 = InvokeLambda(address)
  log.Debug("After executing ", address)
  log.Debug("Latency returned is: ", retval)
  if retval != -1.0 { // Successful invocation
    getDuration(address, retval)
  }

	//SayHello(address, workflowIDs[endpoint])
  
}

// LatencySlice is a thread-safe slice to hold a slice of latency measurements.
type LatencySlice struct {
	sync.Mutex
	slice []float64
}
type LatencySlice_previous struct {
	sync.Mutex
	slice []int64
}

func startMeasurement(msg string) (string, time.Time) {
	return msg, time.Now()
}

func getDuration(msg string, latency float64) {
	//log.Debugf("Invoked %v in %v sec\n", msg, latency)
	log.Infof("Invoked and Executed %v in %v sec\n", msg, latency)
	addDurations(latency)
}

func addDurations(ds float64) {
	latSlice.Lock()
	latSlice.slice = append(latSlice.slice, ds)
	latSlice.Unlock()
}

func getDuration_previous(msg string, start time.Time) {
	latency := time.Since(start)
	log.Debugf("Invoked %v in %v usec\n", msg, latency.Microseconds())
	//addDurations_previous([]time.Duration{latency})
}

/*
func addDurations_previous(ds []time.Duration) {
	latSlice.Lock()
	for _, d := range ds {
		latSlice.slice = append(latSlice.slice, d.Microseconds())
	}
	latSlice.Unlock()
}*/

func writeLatencies(rps float64, latencyOutputFile string) {
	latSlice.Lock()
	defer latSlice.Unlock()

	fileName := fmt.Sprintf("rps%.2f_%s", rps, latencyOutputFile)
	log.Info("The measured latencies are saved in ", fileName)

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		log.Fatal("Failed creating file: ", err)
	}

	datawriter := bufio.NewWriter(file)

	for _, lat := range latSlice.slice {
		//_, err := datawriter.WriteString(strconv.FormatInt(lat, 10) + "\n")
		//_, err := datawriter.WriteString(strconv.FormatFloat(lat, 'E', -1, 64) + "\n")
		_, err := datawriter.WriteString(strconv.FormatFloat(lat, 'f', 14, 64) + "\n")
		if err != nil {
			log.Fatal("Failed to write the URLs to a file ", err)
		}
	}

	datawriter.Flush()
	file.Close()
}
