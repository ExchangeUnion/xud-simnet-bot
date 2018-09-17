package main

import (
	"context"
	"flag"
	"io"
	"net/url"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"github.com/exchangeunion/xud-tests/bot/utils/stackdriver"
	"github.com/exchangeunion/xud-tests/bot/xudrpc"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	nodeAddr1 = flag.String("node1", "02b66438730d1fcdf4a4ae5d3d73e847a272f160fee2938e132b52cab0a0d9cfc6@xud1.test.exchangeunion.com:8885", "XUD node1 address in the format of pubkey@host:port.")
	nodeAddr2 = flag.String("node2", "028599d05b18c0c3f8028915a17d603416f7276c822b6b2d20e71a3502bd0f9e0a@xud2.test.exchangeunion.com:8885", "XUD node2 address in the format of pubkey@host:port.")
	nodeAddr3 = flag.String("node3", "03fd337659e99e628d0487e4f87acf93e353db06f754dccc402f2de1b857a319d0@xud3.test.exchangeunion.com:8885", "XUD node3 address in the format of pubkey@host:port.")
)

func main() {
	println(`
	  ExchangeUnion Trading Tests Bot ====> https://exchangeunion.com/
	`)
	logFormat := new(log.TextFormatter)
	logFormat.TimestampFormat = "2006-01-02 15:04:05"
	logFormat.FullTimestamp = true
	log.SetFormatter(logFormat)
	projectID, err := metadata.ProjectID()
	if err == nil {
		ctx := context.Background()
		stackdriverlogs, err := logging.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create stackdriver logging client: %v", err)
		}
		h := stackdriver.New(stackdriverlogs, "xud-trading-bot")
		log.AddHook(h)
	}

	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	nodegRPCURI, err := url.Parse("//" + *nodeAddr1)
	checkErr(err)
	conn, err := grpc.Dial(nodegRPCURI.Hostname()+":8886", opts...)
	if err != nil {
		log.Fatalf("Failed to connect with node: %v", err)
	}
	ctx := context.Background()
	node := xudrpc.NewXudClient(conn)

	log.Println("Trying to Get Node Info ---> GetInfo() \n")

	nodeoneinfo, err := node.GetInfo(ctx, &xudrpc.GetInfoRequest{})
	checkErr(err)
	log.Warningln("Connected Node Info: ")
	log.Infoln("Node Version: ", nodeoneinfo.Version)
	log.Infoln("Node PubKey: ", nodeoneinfo.NodePubKey)

	log.Println("Asking node to connect with other nodes ---> Connect() \n")
	conres, err := node.Connect(ctx, &xudrpc.ConnectRequest{NodeUri: *nodeAddr2})
	sts, ok := status.FromError(err)
	if !ok && sts.Code().String() != "AlreadyExists" {
		log.Fatalln(sts.Message())
	}
	if conres != nil {
		log.Println(conres)
	} else {
		log.Warningln("Nodes are connected to each other successfully! \n")
	}
	conres, err = node.Connect(ctx, &xudrpc.ConnectRequest{NodeUri: *nodeAddr3})
	sts, ok = status.FromError(err)
	if !ok && sts.Code().String() != "AlreadyExists" {
		log.Fatalln(sts.Message())
	}
	if conres != nil {
		log.Println(conres)
	} else {
		log.Warningln("Nodes  are connected to each other successfully! \n")
	}
	//Listen to PeerOrder & Swap Streams
	go listenPeerOrders(node)
	go listenSwaps(node)

	log.Infoln("Starting Test Trades \n")
	//Indefinite
	for {
		log.Infoln("Checking orders on connected nodes: \n")
		nodeOrders, err := node.GetOrders(ctx, &xudrpc.GetOrdersRequest{PairId: "LTC/BTC", IncludeOwnOrders: true})
		checkErr(err)
		log.Println(nodeOrders)
		orders := nodeOrders.GetOrders()
		//If only one buy order place one more
		if len(orders["LTC/BTC"].BuyOrders) <= 1 {
			log.Infoln("Placing a buy order \n")
			buyOrder, err := node.PlaceOrder(ctx, &xudrpc.PlaceOrderRequest{Price: 200, PairId: "LTC/BTC", Quantity: 0.001, OrderId: uuid.NewV1().String()})
			checkErr(err)
			log.Println(buyOrder)
			println()
		}

		//If only one sell order place one more
		if len(orders["LTC/BTC"].GetSellOrders()) <= 1 {
			log.Infoln("Placing a sell order order \n")
			sellOrder, err := node.PlaceOrder(ctx, &xudrpc.PlaceOrderRequest{Price: 2000, PairId: "LTC/BTC", Quantity: -0.01, OrderId: uuid.NewV1().String()})
			checkErr(err)
			log.Println(sellOrder)
			println()
		}
		// Cancel the order if the order is not fullfilled in the last 24hrs
		cancelOldOrders(ctx, node, orders["LTC/BTC"].GetSellOrders())
		// Cancel the order if the order is not fullfilled in the last 24hrs
		cancelOldOrders(ctx, node, orders["LTC/BTC"].GetBuyOrders())
		time.Sleep(time.Second * 20)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Warningln(err)
	}
}

func cancelOldOrders(ctx context.Context, node xudrpc.XudClient, orders []*xudrpc.Order) {
	for _, order := range orders {
		createdAt := time.Unix(order.CreatedAt, 0)
		if (time.Now().Sub(createdAt).Hours() / 24) > 1 {
			log.Infoln("Cancel the last order with ID: " + order.GetId() + "\n")
			_, err := node.CancelOrder(ctx, &xudrpc.CancelOrderRequest{OrderId: order.Id})
			checkErr(err)
			if err == nil {
				log.Println("Order: " + order.Id + " Successfully cancelled!")
			} else {
				log.Warningln("Order: " + order.Id + " couldn't be cancelled!")
			}
		}
	}
}

func checkMatches(order xudrpc.PlaceOrderResponse) {
	if order.Matches != nil {
		log.Println("We have some order matches: \n")
		log.Println(order.Matches)
	}
	if order.RemainingOrder != nil {
		log.Println("Remaining Order Quantity: \n")
		log.Println(order.RemainingOrder)
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
			log.Warnf("%v.SubscribePeerOrders(_) = _, %v", node, err)
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
			log.Warnf("%v.SubscribeSwaps(_) = _, %v", node, err)
		}
		log.Warningln("Looks like we have a swap event: \n")
		log.Println(swapevent)
	}
}
