package main
import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"
	"os/signal"
	"time"
	"encoding/json"
	"github.com/influxdata/influxdb/client/v2"
)
const (
	MyDB = "readings"
	username = "influkser"
	password = "cukierLukierGancPomada"
)
//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

var dfh MQTT.ConnectionLostHandler = func(client MQTT.Client , err error) {
	fmt.Printf("Connection Lost: %s\n", err)
}

type TempHumidReading struct {
	Temp   float64
	Hum float64 `json:",omitempty"`
	Level int16 `json:",omitempty"`
	PersonCount int16 `json:",omitempty"`
	TStamp string //time.Time, 2016-01-16T10:32:01Z
	RoomName string
}
//TODO make wildcard for a conf room and a type of an entry
//wsdc/szkolna/air -> wsdc/Nick
var topic = "wsdc"

func DeserializeJson(bytearr []byte) (TempHumidReading, error){
	var r TempHumidReading
	unmarshalError := json.Unmarshal(bytearr, &r)
	if (unmarshalError != nil) {
		fmt.Println("Cannot parse ", string(bytearr), "to json:",unmarshalError)
		return r, unmarshalError
	}
	fmt.Println("Deserialize:", string(bytearr))
	fmt.Printf("Deserialized: %+v", r)
	return r, nil
}
func getMeTheObject(r TempHumidReading){
	if(r.Temp!=nil && r.Hum!=nil){
		return map[string]interface{}{
			"temp":   r.Temp,
			"hum": 	r.Hum,
		}
	}
	if(r.Level!=nil){
		return map[string]interface{}{
			"lvl": 	r.Level,
		}
	}
	if(r.PersonCount!=nil){
		return map[string]interface{}{
			"personCount": r.PersonCount,
		}
	}
}
func main() {
	hostname, _ := os.Hostname()
	fmt.Println(hostname)

	// Make db client  // was: 192.168.1.108
	dbClient, _ := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://0.0.0.0:8086",
		Username: username,
		Password: password,
	})
	fmt.Println("initiated influx db client")
	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MyDB,
		Precision: "s",
	})

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler //was: 192.168.1.108
	opts := MQTT.NewClientOptions().AddBroker("tcp://0.0.0.0:1883")
	opts.SetKeepAlive(0) //500*time.Second
	opts.SetClientID("go-influks")
	opts.SetDefaultPublishHandler(f)
	opts.SetConnectionLostHandler(dfh)
	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	fmt.Println("started MQTT client with KeepAlive : ", opts.KeepAlive, " AutoReconnect: ", opts.AutoReconnect, " Clean Session:",opts.CleanSession)

	if connectoken := c.Connect(); connectoken.Wait() && connectoken.Error() != nil {
		fmt.Println("connect tokken err")
		panic(connectoken.Error())
	}

	//subscribe to the topic /go-mqtt/sample and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription

	if token := c.Subscribe(topic, 0, func(mqttClient MQTT.Client, msg MQTT.Message) {
		fmt.Println("Received: Topic=", msg.Topic())
		r,deserializeErr := DeserializeJson(msg.Payload())
		if(deserializeErr!= nil){
			fmt.Println("json deserialization issue:", deserializeErr, "topic:", msg.Topic())
			return
		}

		// Create a point and add to batch
		tags := map[string]string{"sensors": r.RoomName}
		fields := getMeTheObject(r)
		//RFC3339
		t, _ := time.Parse("2015-12-22T01:17:39Z", r.TStamp)
		pt,err := client.NewPoint(r.RoomName, tags, fields, t)
		if(err != nil){
			fmt.Println("Cannot create time series point", tags, fields, r.TStamp)
		}else{
			bp.AddPoint(pt)
			// Write the batch
			dbClient.Write(bp)
		}
	}); token.Wait() && token.Error() != nil {
		fmt.Println("MQTT subscription error while subscribing to topic", topic)
		fmt.Println(token.Error())
		os.Exit(1)
	}

	//Publish 5 messages to /go-mqtt/sample at qos 1 and wait for the receipt
	//    token := c.Publish("w112/sensors/temperature/kitchen", 0, false, text)
	
	fmt.Println("Interrupt to exit (Ctrl+C)")
	channello := make(chan os.Signal, 1)
	signal.Notify(channello, os.Interrupt)
	go func(){
	    for sig := range channello {
	        // sig is a ^C, handle it
	        fmt.Println("exiting aafter interrupt signal", sig)
	        if unsubToken := c.Unsubscribe(topic); unsubToken.Wait() && unsubToken.Error() != nil {
				fmt.Println(unsubToken.Error())
				os.Exit(1)
			}
			c.Disconnect(250)
			os.Exit(0)
		}
	}()

	for {
		time.Sleep(30000 * time.Millisecond)
		fmt.Print("\xe2\x99\xa5")
	}
}
