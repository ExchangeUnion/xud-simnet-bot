package trading

import (
	"sync"
	"time"

	"github.com/ExchangeUnion/xud-tests/xudclient"
	"github.com/ExchangeUnion/xud-tests/xudrpc"
)

var xud *xudclient.Xud
var mode string
//var swaps = 0

var openOrders = make(map[string]*openOrder)

type placeOrderParameters struct {
	price    float64
	quantity uint64
	side     xudrpc.OrderSide
}

type openOrder struct {
	quantityLeft uint64

	// What should be placed once the order is filled completely
	toPlace placeOrderParameters
}

// InitTradingBot initializes a new trading bot
func InitTradingBot(wg *sync.WaitGroup, xudclient *xudclient.Xud, tradingMode string) {
	xud = xudclient
	mode = tradingMode

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {

			// ensure connectivity to xud
			info, err := xud.GetInfo()
			for err != nil {
				log.Error("Lost connection to XUD. Retrying in 5 seconds")
				time.Sleep(5 * time.Second)
				xud.Init()
				info, err = xud.GetInfo()
			}
			log.Infof("%v", info)

			log.Debug("Placing orders")

			go placeOrders()
			log.Debug("Subscribing to order events")

			go subscribeAddedOrders()
			subscribeRemovedOrders()
		}

	}()
}

func subscribeAddedOrders() error {
	err := xud.SubscribeAddedOrders(orderAdded)
	return err
}


func subscribeRemovedOrders() error {
	err := xud.SubscribeRemovedOrders(orderRemoved)
	log.Infof("SubscribeRemovedOrders: exiting %v", err)
	return nil
}

func placeOrders() error {

	orders := []placeOrderParameters{
		{
			price:    0.0077,
			quantity: 3.5*1E8,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0076,
			quantity: 15.5*1E8,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0075,
			quantity: 18.0*1E8,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0074,
			quantity: 21.25*1E8,
			side:     xudrpc.OrderSide_BUY,
		},
		{
			price:    0.0073,
			quantity: 24.0*1E8,
			side:     xudrpc.OrderSide_BUY,
		},

		{
			price:    0.0079,
			quantity: 2.5*1E8,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0080,
			quantity: 13*1E8,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0081,
			quantity: 15.6*1E8,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0082,
			quantity: 18.1*1E8,
			side:     xudrpc.OrderSide_SELL,
		},
		{
			price:    0.0083,
			quantity: 22.3*1E8,
			side:     xudrpc.OrderSide_SELL,
		},
	}

	switch  mode {
	case "standard":
		for _, order := range orders {
			err := placeOrder(order)
			if err != nil {
				log.Errorf("Could not place orders: %v - %v", order, err)
				return err
			}
		}
		log.Debug("Placed orders")
		break
	case "2.5@0.0079":
		order := placeOrderParameters{
			price:    0.0079,
			quantity: 2.5*1E8,
			side:     xudrpc.OrderSide_BUY,
		}
		err := fillOrKill(order)
		if err != nil {
			log.Errorf("Could not place FOK order: %v - %v", order, err)
			return err
		}
		break
	case "3.5@0.0077":
		order := placeOrderParameters{
			price:    0.0077,
			quantity: 3.5*1E8,
			side:     xudrpc.OrderSide_SELL,
		}
		err := fillOrKill(order)
		if err != nil {
			log.Errorf("Could not place FOK order: %v - %v", order, err)
			return err
		}
		break
	default:

	}

	return nil
}

func placeOrder(params placeOrderParameters) error {
	req := xudrpc.PlaceOrderRequest{
		Price:    params.price,
		Quantity: params.quantity,
		Side:     params.side,
		PairId:   "LTC/BTC",
	}
	log.Debugf("Placing order: %v ", req)
	response, err := xud.PlaceOrderSync(req)


	if err != nil {
		return err
	}

	//if response.SwapResults != nil {
	//	log.Debugf("Swapped: %v ", response.SwapResults)
	//}

	var remainingOrder = response.RemainingOrder


	// check - why this should be done? maybe the stream opens after the orders?
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

func fillOrKill(params placeOrderParameters) error {
	req := xudrpc.PlaceOrderRequest{
		Price:    params.price,
		Quantity: params.quantity,
		Side:     params.side,
		PairId:   "LTC/BTC",
	}
	log.Debugf("Placing FOK order: %v ", req)
	response, err := xud.PlaceOrderSync(req)

	if err != nil {
		log.Errorf("FOK place order failed: %v", err)
		return err
	}

	//if response.SwapResults != nil {
	//	swaps++
	//	log.Debugf("#%v FOK Swapped: %v ", swaps, response.SwapResults)
	//}

	if (response.RemainingOrder != nil) {
		killReq := &xudrpc.RemoveOrderRequest{
			OrderId:	response.RemainingOrder.Id,
		}
		resp, err := xud.Client.RemoveOrder(xud.Ctx, killReq)
		if (err != nil) {
			log.Errorf("Failed to remove Remaining order : %v", resp.String())
			return err
		}
		log.Debugf("Remaining order removed: %v", resp.String())

	}
	return nil
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


func orderAdded(newOrder xudrpc.Order) {
	switch mode {
	case "standard":
		log.Debug("Order added: %v", newOrder)
		break
	case "2.5@0.0079":
		if newOrder.Price == 0.0079 && newOrder.Quantity == 2.5*1E8 && !newOrder.IsOwnOrder && newOrder.Side == xudrpc.OrderSide_SELL {
			log.Debug("Order detected: %v", newOrder)
			order := placeOrderParameters{
				price:    0.0079,
				quantity: 2.5*1E8,
				side:     xudrpc.OrderSide_BUY,
			}
			err := fillOrKill(order)
			if err != nil {
				log.Errorf("Could not place FOK order: %v - %v", order, err)
				return
			}
			break
		}
	case "3.5@0.0077":
		if newOrder.Price == 0.0077 && newOrder.Quantity == 3.5*1E8 && !newOrder.IsOwnOrder && newOrder.Side == xudrpc.OrderSide_BUY {
			log.Debug("Order detected: %v", newOrder)
			order := placeOrderParameters{
				price:    0.0077,
				quantity: 3.5*1E8,
				side:     xudrpc.OrderSide_SELL,
			}
			err := fillOrKill(order)
			if err != nil {
				log.Errorf("Could not place FOK order: %v - %v", order, err)
				return
			}
			break
		}

	}



}