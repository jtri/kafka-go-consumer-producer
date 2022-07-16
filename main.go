package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Shopify/sarama"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kgcp",
		Short: "kgpc (kafka-go-consumer-producer) doubles as a kafka consumer or producer",
	}
	consumerCmd = &cobra.Command{
		Use:   "consumer",
		Short: "start consumer",
		RunE:  consumerRunE,
	}
	producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "start producer",
		RunE:  producerRunE,
	}

	// kafka is configured with auto.create.topics.enable=true meaning non-existent topic are created
	// when an application attempts to consume or produce against that topic name
	topic = "my_topic"
	addrs = []string{"localhost:9092"}
)

func init() {
	rootCmd.AddCommand(consumerCmd)
	rootCmd.AddCommand(producerCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// normally this would be defined in protocol buffers and we would import the generated code
type OurFancyMessage struct {
	Timestamp time.Time
}

func (o *OurFancyMessage) String() string {
	return fmt.Sprintf("our-fancy-message-%v", o.Timestamp)
}

func consumerRunE(cmd *cobra.Command, args []string) error {
	consumer, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	consumed := 0
ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Consumed message offset %d\n", msg.Offset)
			log.Printf("Consumed message value %s\n", msg.Value)
			consumed++
		case <-interrupt:
			break ConsumerLoop
		}
	}

	log.Printf("Consumed: %d\n", consumed)
	return nil
}

func producerRunE(cmd *cobra.Command, args []string) error {
	producer, err := sarama.NewSyncProducer(addrs, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-interrupt
		cancel()
	}()

	ticker := time.NewTicker(time.Second * 15)

	for {
		ofm := &OurFancyMessage{Timestamp: time.Now()}
		msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(ofm.String())}
		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("FAILED to send message: %s\n", err)
		} else {
			log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
		}
		select {
		case <-ticker.C:
			log.Println("producing next message")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
