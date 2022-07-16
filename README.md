# kafka-go-consumer-producer

Simple app to play with kafka locally as a producer and consumer using the sarama library.

This repo assumes you have a running kubernetes cluster already.

## Deploy kafka

        kubectl apply -f kafka.yaml

We are using the Kraft feature of Kafka to get rid of the zookeeper dependency and make our development environment simpler. Kafka is also configured with `auto.create.topics.enable=true` so
we do not need to worry about bootstrapping.

## Interacting with Kafka via Producer or Consumer

We port-forward the kafka broker instead of figuring out ingress on a kubernetes cluster

        kubectl port-forward svc/kafka 9092:9092

After we make Kafka available on localhost, we can run the producer and consumer in separate terminals

        go run main.go consumer # run the consumer

        go run main.go producer # run the producer

