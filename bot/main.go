package main

import (
	"context"
	"flag"
	"io"
	"net/url"
	"time"

	"github.com/exchangeunion/xud-tests/bot/xudrpc"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	nodeAddr1 = flag.String("node1", "02b66438730d1fcdf4a4ae5d3d73e847a272f160fee2938e132b52cab0a0d9cfc6@xud1.test.exchangeunion.com:8885", "XUD node address in the format of pubkey@host:port.")
	nodeAddr2 = flag.String("node2", "028599d05b18c0c3f8028915a17d603416f7276c822b6b2d20e71a3502bd0f9e0a@xud2.test.exchangeunion.com:8885", "XUD node address in the format of pubkey@host:port.")
)

func main() {
	println(`
	ExchangeUnion Test Bot ====>
	https://exchangeunion.com/
	`)
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	nodeOnegRPCURI, err := url.Parse("//" + *nodeAddr1)
	checkErr(err)
	nodeTwogRPCURI, err := url.Parse("//" + *nodeAddr2)
	checkErr(err)
	conn1, err := grpc.Dial(nodeOnegRPCURI.Hostname()+":8886", opts...)
	if err != nil {
		log.Fatalf("failed to connect with node1: %v", err)
	}
	defer conn1.Close()
	conn2, err := grpc.Dial(nodeTwogRPCURI.Hostname()+":8886", opts...)
	if err != nil {
		log.Fatalf("failed to connect with node2: %v", err)
	}
	defer conn2.Close()
	ctx := context.Background()
	node1 := xudrpc.NewXudClient(conn1)
	node2 := xudrpc.NewXudClient(conn2)
	log.Println("Trying to Get Nodes Info ---> GetInfo() \n")

	nodeoneinfo, err := node1.GetInfo(ctx, &xudrpc.GetInfoRequest{})
	checkErr(err)
	log.Warningln("Node1 Info: ")
	log.Infoln("Node Version: ", nodeoneinfo.Version)
	log.Infoln("Node PubKey: ", nodeoneinfo.NodePubKey)

	nodetwoinfo, err := node2.GetInfo(ctx, &xudrpc.GetInfoRequest{})
	checkErr(err)
	log.Warningln("Node2 Info:")
	log.Infoln("Node Version: ", nodetwoinfo.Version)
	log.Infoln("Node PubKey: ", nodetwoinfo.NodePubKey)

	log.Println("Asking nodes to connect with each other ---> Connect() \n")
	conres, err := node1.Connect(ctx, &xudrpc.ConnectRequest{NodeUri: *nodeAddr2})
	sts, ok := status.FromError(err)
	if !ok && sts.Code().String() != "AlreadyExists" {
		log.Fatalln(sts.Message())
	}
	if conres != nil {
		log.Println(conres)
	} else {
		log.Warningln("Nodes Connected to eachother successfully! \n")
	}
	//Listen to Streams
	go listenPeerOrders(node1)
	go listenPeerOrders(node1)
	go listenSwaps(node1)
	go listenSwaps(node2)

	log.Infoln("Starting Test Trades \n")
	//Indefinite
	for {
		//TODO implement market orders
		log.Infoln("Placing some test orders \n")
		firstOrder, err := node1.PlaceOrder(ctx, &xudrpc.PlaceOrderRequest{Price: 2000, PairId: "BTC/LTC", Quantity: -50, OrderId: uuid.NewV1().String()})
		checkErr(err)
		log.Println(firstOrder)
		thirdOrder, err := node2.PlaceOrder(ctx, &xudrpc.PlaceOrderRequest{Price: 2000, PairId: "BTC/LTC", Quantity: 50, OrderId: uuid.NewV1().String()})
		checkErr(err)
		if thirdOrder.Matches != nil {
			log.Println("We have some order matches: \n")
			log.Println(thirdOrder.Matches)
		}
		if thirdOrder.RemainingOrder != nil {
			log.Println("Remaining Order Quantity: \n")
			log.Println(thirdOrder.RemainingOrder)
		}
		secondOrder, err := node1.PlaceOrder(ctx, &xudrpc.PlaceOrderRequest{Price: 2000, PairId: "BTC/LTC", Quantity: -50, OrderId: uuid.NewV1().String()})
		checkErr(err)
		log.Println(secondOrder)
		log.Println("")
		log.Infoln("Cancel the last order with ID: " + secondOrder.RemainingOrder.GetId() + "\n")
		cancelOrder, err := node1.CancelOrder(ctx, &xudrpc.CancelOrderRequest{OrderId: secondOrder.RemainingOrder.Id, PairId: "BTC/LTC"})
		checkErr(err)
		if cancelOrder.Canceled {
			log.Println("Order: " + secondOrder.RemainingOrder.Id + " Successfully cancelled!")
		} else {
			log.Warningln("Order: " + secondOrder.RemainingOrder.Id + " couldn't be cancelled!")
		}
		log.Infoln("Checking orders on connected nodes: \n")
		nodeOneOrders, err := node1.GetOrders(ctx, &xudrpc.GetOrdersRequest{PairId: "BTC/LTC"})
		checkErr(err)
		log.Println(nodeOneOrders)
		nodeTwoOrders, err := node2.GetOrders(ctx, &xudrpc.GetOrdersRequest{PairId: "BTC/LTC"})
		checkErr(err)
		log.Println(nodeTwoOrders)
		time.Sleep(1000)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func listenPeerOrders(node xudrpc.XudClient) {
	log.Infoln("Starting listening to PeerOrders")
	orderstream, err := node.SubscribePeerOrders(context.Background(), &xudrpc.SubscribePeerOrdersRequest{})
	checkErr(err)
	for {
		peerOrder, err := orderstream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.SubscribePeerOrders(_) = _, %v", node, err)
		}
		log.Warningln("Looks like we have a new peer order: \n")
		log.Println(peerOrder)
	}
}

func listenSwaps(node xudrpc.XudClient) {
	log.Infoln("Starting listening to SwapEvents")
	swapstream, err := node.SubscribeSwaps(context.Background(), &xudrpc.SubscribeSwapsRequest{})
	checkErr(err)
	for {
		swapevent, err := swapstream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.SubscribeSwaps(_) = _, %v", node, err)
		}
		log.Warningln("Looks like we have a swap event: \n")
		log.Println(swapevent)
	}
}
