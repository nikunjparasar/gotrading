package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

// equals
func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v !=  %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(BUY, 5)
	buyOrderB := NewOrder(BUY, 8)
	buyOrderC := NewOrder(BUY, 10)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	fmt.Println(l)

}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(SELL, 10)
	sellOrderB := NewOrder(SELL, 5)
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(9_000, sellOrderB)

	assert(t, len(ob.Asks().Limits), 2)

	for _, limit := range ob.Asks().Limits {
		fmt.Println(limit.totalvol)
	}

}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(SELL, 20)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(BUY, 10)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.Asks().Limits), 1)
	assert(t, ob.AsksTotalVolume(), 10.0)
	assert(t, matches[0].ask, sellOrder)
	assert(t, matches[0].bid, buyOrder)
	assert(t, matches[0].sizefilled, 10.0)
	assert(t, matches[0].price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)

	fmt.Printf("%+v", matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(BUY, 5)
	buyOrderB := NewOrder(BUY, 8)
	buyOrderC := NewOrder(BUY, 10)
	buyOrderD := NewOrder(BUY, 1)

	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	assert(t, ob.BidsTotalVolume(), 24.0)

	sellOrder := NewOrder(SELL, 20)
	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, ob.BidsTotalVolume(), 4.0)
	assert(t, len(matches), 3)
	assert(t, len(ob.Bids().Limits), 1)

	fmt.Printf("%+v", matches)

}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(BUY, 4)
	ob.PlaceLimitOrder(10_000, buyOrder)

	assert(t, ob.BidsTotalVolume(), 4.0)

	ob.CancelOrder(buyOrder)

	assert(t, ob.BidsTotalVolume(), 0.0)

}
