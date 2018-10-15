package trading

import (
	"sync"

	"github.com/ExchangeUnion/xud-tests/xudclient"
	"github.com/ExchangeUnion/xud-tests/xudrpc"
)

var xud *xudclient.Xud

var openOrders = make(map[string]*openOrder)
var openOrdersLock = sync.RWMutex{}

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

		// TODO: handle XUD getting down
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

	// Add each order five times
	for _, order := range orders {
		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func(order placeOrderParameters) {
				placeOrder(order)
				wg.Done()
			}(order)
		}
	}

	wg.Wait()

	log.Debug("Placed orders")
}

func placeOrder(params placeOrderParameters) {
	response, err := xud.PlaceOrderSync(xudrpc.PlaceOrderRequest{
		Price:    params.price,
		Quantity: params.quantity,
		Side:     params.side,
		PairId:   "LTC/BTC",
	})

	if err != nil {
		log.Error("Could not place order: %v", err)
		return
	}

	var remainingOrder = response.RemainingOrder

	// Place a new order until there is quantity remaining
	if remainingOrder == nil || remainingOrder.Quantity == 0 {
		log.Debug("Nothing left of placed order: placing new one")
		placeOrder(params)

		return
	}

	openOrdersLock.Lock()

	openOrders[remainingOrder.Id] = &openOrder{
		quantityLeft: remainingOrder.Quantity,
		toPlace:      params,
	}

	openOrdersLock.Unlock()
}

func orderRemoved(removal xudrpc.OrderRemoval) {
	log.Debug("Order removed: %v", removal)

	openOrdersLock.RLock()

	filledOrder := openOrders[removal.OrderId]

	openOrdersLock.RUnlock()

	if filledOrder != nil {
		filledOrder.quantityLeft -= removal.Quantity

		// Check if there is quantity left and place new order if not
		if filledOrder.quantityLeft == 0 {
			log.Debug("Placing new order")
			placeOrder(filledOrder.toPlace)
		}
	}
}
