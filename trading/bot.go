package trading

import (
	"sync"

	"github.com/ExchangeUnion/xud-tests/xudclient"
	"github.com/ExchangeUnion/xud-tests/xudrpc"
)

var xud *xudclient.Xud

type placeOrderParameters struct {
	price    float64
	quantity float64
	side     xudrpc.OrderSide
}

// InitTradingBot initialized a new trading bot
func InitTradingBot(xudclient *xudclient.Xud) {
	xud = xudclient

	go func() {
		log.Debug("Subscribing to removed orders")

		err := xud.SubscribeRemovedOrders(orderRemoved)

		if err != nil {
			log.Error("Lost connection to XUD")
		}
	}()

	placeOrders()
}

func placeOrders() {
	var wg sync.WaitGroup

	orders := []placeOrderParameters{
		{
			price:    0.001,
			quantity: 0.01,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.002,
			quantity: 0.01,
			side:     xudrpc.OrderSide_SELL,
		},
	}

	// Add each other five times
	for _, order := range orders {
		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func(order placeOrderParameters) {
				defer wg.Done()
				placeOrder(order)
			}(order)
		}
	}

	wg.Wait()

	log.Debug("Placed orders")
}

func placeOrder(params placeOrderParameters) *xudrpc.PlaceOrderResponse {
	response, err := xud.PlaceOrder(xudrpc.PlaceOrderRequest{
		Price:    params.price,
		Quantity: params.quantity,
		Side:     params.side,
		PairId:   "LTC/BTC",
	})

	if err != nil {
		log.Error("Could not place order: %v", err)
	}

	return response
}

func orderRemoved(removal xudrpc.OrderRemoval) {
	log.Debug("Order removed, %v", removal)
}
