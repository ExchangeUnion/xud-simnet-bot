package trading

import (
	"sync"
	"time"

	"github.com/ExchangeUnion/xud-tests/xudclient"
	"github.com/ExchangeUnion/xud-tests/xudrpc"
)

var xud *xudclient.Xud

var openOrders = make(map[string]*openOrder)

type placeOrderParameters struct {
	price    float64
	quantity float64
	side     xudrpc.OrderSide
}

type openOrder struct {
	quantityLeft float64

	// What should be placed once the order is filled completely
	toPlace placeOrderParameters
}

// InitTradingBot initializes a new trading bot
func InitTradingBot(wg *sync.WaitGroup, xudclient *xudclient.Xud) {
	xud = xudclient

	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Debug("Subscribing to removed orders")

		err := startXudSubscription()

		for err != nil {
			openOrders = make(map[string]*openOrder)

			log.Error("Lost connection to XUD. Retrying in 5 seconds")
			time.Sleep(5 * time.Second)

			startXudSubscription()
		}
	}()
}

func startXudSubscription() error {
	err := placeOrders()

	if err != nil {
		return err
	}

	err = xud.SubscribeRemovedOrders(orderRemoved)

	if err != nil {
		return err
	}

	if len(openOrders) != 0 {
		// TODO: check if the orders still exist
	}

	return nil
}

func placeOrders() error {

	orders := []placeOrderParameters{
		{
			price:    0.0077,
			quantity: 13.0,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0076,
			quantity: 15.5,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0075,
			quantity: 18.0,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0074,
			quantity: 21.25,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0073,
			quantity: 24.0,
			side:     xudrpc.OrderSide_BUY,
		},

		{
			price:    0.0079,
			quantity: 11.5,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0080,
			quantity: 13,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0081,
			quantity: 15.6,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0082,
			quantity: 18.1,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0083,
			quantity: 22.3,
			side:     xudrpc.OrderSide_SELL,
		},
	}

	var err error

	for _, order := range orders {
			placeErr := placeOrder(order)
			if placeErr != nil {
				err = placeErr
			}
	}

	if err != nil {
		log.Warning("Could not place orders: %v", err)
		return err
	}

	log.Debug("Placed orders")
	return nil
}

func placeOrder(params placeOrderParameters) error {
	response, err := xud.PlaceOrderSync(xudrpc.PlaceOrderRequest{
		Price:    params.price,
		Quantity: params.quantity,
		Side:     params.side,
		PairId:   "LTC/BTC",
	})

	if err != nil {
		return err
	}

	var remainingOrder = response.RemainingOrder

	// Place a new order until there is quantity remaining
	if remainingOrder == nil || remainingOrder.Quantity == 0 {
		log.Debug("Nothing left of placed order: placing new one")
		err = placeOrder(params)

		return err
	}

	openOrders[remainingOrder.Id] = &openOrder{
		quantityLeft: remainingOrder.Quantity,
		toPlace:      params,
	}

	return err
}

func orderRemoved(removal xudrpc.OrderRemoval) {
	log.Debug("Order removed: %v", removal)

	filledOrder := openOrders[removal.OrderId]

	if filledOrder != nil {
		filledOrder.quantityLeft -= removal.Quantity

		// Check if there is quantity left and place new order if not
		if filledOrder.quantityLeft == 0 {
			log.Debug("Placing new order: %v", filledOrder.toPlace)
			placeOrder(filledOrder.toPlace)
		}
	}
}
