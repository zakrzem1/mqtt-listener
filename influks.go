package main
import (
	"fmt"
//import the Paho Go MQTT library
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"os"
	"time"
	"io"
	"bufio"
//	"net/url"
//	"log"
	"encoding/json"
	"github.com/influxdb/influxdb/client/v2"
)
const (
	MyDB = "readings"
	username = "influkser"
	password = "cukierLukierGancPomada"
)
//define a funct  ion for the default message handler
var f MQTT.MessageHandler = func(client *MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

type TempReading struct {
	temp   float64
	tstamp string //time.Time
}

var topic = "w112/sensors/temperature/kitchen"

func DeserializeJson(bytearr []byte) (TempReading, error){
	var r TempReading
	unmarshalError := json.Unmarshal(bytearr, &r)
	if (unmarshalError != nil) {
		fmt.Println("Cannot parse ", string(bytearr), "to json:",unmarshalError)
		return nil, unmarshalError
	}
	fmt.Println("Deserialize:", string(bytearr))
	fmt.Printf("Deserialized: %+v", r)
	return r, nil
}
func main() {
	stdin := bufio.NewReader(os.Stdin)
	hostname, _ := os.Hostname()
	fmt.Println(hostname)

	// Make db client
	dbClient, _ := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
		Username: username,
		Password: password,
	})

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MyDB,
		Precision: "s",
	})

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := MQTT.NewClientOptions().AddBroker("tcp://192.168.1.107:1883")
	opts.SetClientID("go-simple")
	opts.SetDefaultPublishHandler(f)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	//subscribe to the topic /go-mqtt/sample and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription

	if token := c.Subscribe(topic, 0, func(mqttClient *MQTT.Client, msg MQTT.Message) {
		fmt.Println("Received: Topic=", msg.Topic())
		r,err := DeserializeJson(msg.Payload())

		// Create a point and add to batch
		tags := map[string]string{"temp": "w112-kuchnia"}
		fields := map[string]interface{}{
			"temp":   r.temp,
		}
		//RFC3339
		t, _ := time.Parse("2015-12-22T01:17:39Z", r.tstamp)
		pt,err := client.NewPoint("w112_temp", tags, fields, t)
		if(err != nil){
			fmt.Println("Cannot create time series point", tags, fields, r.tstamp)
		}else{
			bp.AddPoint(pt)
			// Write the batch
			dbClient.Write(bp)
		}
	}); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	//Publish 5 messages to /go-mqtt/sample at qos 1 and wait for the receipt
	//from the server after sending each message
	//  for i := 0; i < 5; i++ {
	//    text := fmt.Sprintf("this is msg #%d!", i)
	//    token := c.Publish("w112/sensors/temperature/kitchen", 0, false, text)
	//    token.Wait()
	//  }
	for {
		message, err := stdin.ReadString('\n')
		print(message)
		if err == io.EOF {
			os.Exit(0)
		}
		//		client.Publish(*topic, byte(*qos), *retained, message)
	}

	//unsubscribe from /go-mqtt/sample
	if token := c.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	c.Disconnect(250)
}
