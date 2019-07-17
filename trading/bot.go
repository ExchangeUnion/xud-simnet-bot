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
	pairID   string
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
		// BTC/DAI orders
		{
			price:    9998,
			quantity: 3.5 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "BTC/DAI",
		},
		{
			price:    9997,
			quantity: 15.5 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "BTC/DAI",
		},
		{
			price:    9996,
			quantity: 18 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "BTC/DAI",
		},
		{
			price:    9995,
			quantity: 21.25 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "BTC/DAI",
		},
		{
			price:    9994,
			quantity: 24 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "BTC/DAI",
		},

		// {
		// 	price:    10000,
		// 	quantity: 2.5 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "BTC/DAI",
		// },
		// {
		// 	price:    10001,
		// 	quantity: 13 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "BTC/DAI",
		// },
		// {
		// 	price:    10002,
		// 	quantity: 15.6 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "BTC/DAI",
		// },
		// {
		// 	price:    10003,
		// 	quantity: 18.1 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "BTC/DAI",
		// },
		// {
		// 	price:    10004,
		// 	quantity: 22.3 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "BTC/DAI",
		// },

		// LTC/DAI orders
		{
			price:    98,
			quantity: 3.5 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "LTC/DAI",
		},
		{
			price:    97,
			quantity: 15.5 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "LTC/DAI",
		},
		{
			price:    96,
			quantity: 18 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "LTC/DAI",
		},
		{
			price:    95,
			quantity: 21.25 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "LTC/DAI",
		},
		{
			price:    94,
			quantity: 24 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "LTC/DAI",
		},

		// {
		// 	price:    100,
		// 	quantity: 2.5 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "LTC/DAI",
		// },
		// {
		// 	price:    101,
		// 	quantity: 13 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "LTC/DAI",
		// },

		// {
		// 	price:    102,
		// 	quantity: 15.6 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "LTC/DAI",
		// },
		// {
		// 	price:    103,
		// 	quantity: 18.1 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "LTC/DAI",
		// },
		// {
		// 	price:    104,
		// 	quantity: 22.3 * 1E8,
		// 	side:     xudrpc.OrderSide_SELL,
		// 	pairID:   "LTC/DAI",
		// },

		// WETH/BTC orders
		// {
		// 	price:    0.021,
		// 	quantity: 3.5 * 1E8,
		// 	side:     xudrpc.OrderSide_BUY,
		// 	pairID:   "WETH/BTC",
		// },
		// {
		// 	price:    0.02,
		// 	quantity: 15.5 * 1E8,
		// 	side:     xudrpc.OrderSide_BUY,
		// 	pairID:   "WETH/BTC",
		// },
		// {
		// 	price:    0.019,
		// 	quantity: 18 * 1E8,
		// 	side:     xudrpc.OrderSide_BUY,
		// 	pairID:   "WETH/BTC",
		// },
		// {
		// 	price:    0.018,
		// 	quantity: 21.25 * 1E8,
		// 	side:     xudrpc.OrderSide_BUY,
		// 	pairID:   "WETH/BTC",
		// },
		// {
		// 	price:    0.017,
		// 	quantity: 24 * 1E8,
		// 	side:     xudrpc.OrderSide_BUY,
		// 	pairID:   "WETH/BTC",
		// },

		{
			price:    0.023,
			quantity: 2.5 * 1E8,
			side:     xudrpc.OrderSide_SELL,
			pairID:   "WETH/BTC",
		},
		{
			price:    0.024,
			quantity: 13 * 1E8,
			side:     xudrpc.OrderSide_SELL,
			pairID:   "WETH/BTC",
		},
		{
			price:    0.025,
			quantity: 15.6 * 1E8,
			side:     xudrpc.OrderSide_SELL,
			pairID:   "WETH/BTC",
		},
		{
			price:    0.026,
			quantity: 18.1 * 1E8,
			side:     xudrpc.OrderSide_SELL,
			pairID:   "WETH/BTC",
		},
		{
			price:    0.027,
			quantity: 22.3 * 1E8,
			side:     xudrpc.OrderSide_SELL,
			pairID:   "WETH/BTC",
		},
	}

	switch mode {
	case "standard":
		for _, order := range orders {
			err := placeOrder(order)
			if err != nil {
				log.Errorf("Could not place orders: %v - %v", order, err)
				//	return err
			}
		}
		log.Debug("Placed orders")
		break
	case "2.5@0.0079":
		order := placeOrderParameters{
			price:    0.0079,
			quantity: 2.5 * 1E8,
			side:     xudrpc.OrderSide_BUY,
			pairID:   "LTC/BTC",
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
			quantity: 3.5 * 1E8,
			side:     xudrpc.OrderSide_SELL,
			pairID:   "LTC/BTC",
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
		PairId:   params.pairID,
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
		PairId:   params.pairID,
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

	if response.RemainingOrder != nil {
		killReq := &xudrpc.RemoveOrderRequest{
			OrderId: response.RemainingOrder.Id,
		}
		resp, err := xud.Client.RemoveOrder(xud.Ctx, killReq)
		if err != nil {
			log.Errorf("Failed to remove Remaining order : %v", resp.String())
			return err
		}
		log.Debugf("Remaining order removed: %v", resp.String())

	}
	return nil
}

func orderRemoved(update xudrpc.OrderUpdate) {
	removal := update.GetOrderRemoval()
	if removal == nil {
		return
	}
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

func orderAdded(update xudrpc.OrderUpdate) {
	newOrder := update.GetOrder()
	if newOrder == nil {
		return
	}
	switch mode {
	case "standard":
		log.Debug("Order added: %v", newOrder)
		break
	case "2.5@0.0079":
		if newOrder.Price == 0.0079 && newOrder.Quantity == 2.5*1E8 && !newOrder.IsOwnOrder && newOrder.Side == xudrpc.OrderSide_SELL {
			log.Debug("Order detected: %v", newOrder)
			order := placeOrderParameters{
				price:    0.0079,
				quantity: 2.5 * 1E8,
				side:     xudrpc.OrderSide_BUY,
				pairID:   "LTC/BTC",
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
				quantity: 3.5 * 1E8,
				side:     xudrpc.OrderSide_SELL,
				pairID:   "LTC/BTC",
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
